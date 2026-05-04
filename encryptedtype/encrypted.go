// Package encryptedtype provides EncryptedString — a SQL column type that
// transparently encrypts plaintext on write and decrypts on read.
//
// On the way out (Value) the plaintext is sealed with AES-256-GCM via the
// configured Sealer from github.com/ubgo/crypt.
//
// On the way in (Scan) the ciphertext is opened via crypt.OpenAuto, which
// transparently dispatches between AES-256-GCM (the modern AEAD format)
// and AES-CBC (a peer format kept first-class for interop). Existing
// CBC-encrypted columns continue to read without a migration step.
//
// Boot wiring is required exactly once per process:
//
//	key, _ := loadEncryptionKey()
//	if err := encryptedtype.SetKey(key); err != nil {
//	    log.Fatal(err)
//	}
//
// gqlgen integration is duck-typed: this package does not import gqlgen.
//
// Defense in depth: String returns "[encrypted]" and MarshalJSON returns
// null so the plaintext does not leak via fmt or JSON paths. The only
// outbound channels that see plaintext are MarshalGQL (because the
// schema author opted in by exposing the field) and Plain() (an explicit
// caller-side accessor).
package encryptedtype

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/ubgo/crypt"
)

// EncryptedString stores plaintext in memory and encrypts/decrypts at
// the SQL boundary.
type EncryptedString struct {
	plain string
	set   bool
}

// New returns an EncryptedString holding the given plaintext.
func New(plaintext string) EncryptedString {
	return EncryptedString{plain: plaintext, set: true}
}

// Plain returns the in-memory plaintext. Callers that hold an
// EncryptedString already have access to the value; this accessor is
// for code that needs to operate on the raw string (e.g. forwarding it
// to a third-party SDK).
func (e EncryptedString) Plain() string { return e.plain }

// IsSet reports whether the value has been assigned.
func (e EncryptedString) IsSet() bool { return e.set }

// Value implements [driver.Valuer]. Unset values persist as SQL NULL.
// Set values are sealed with AES-256-GCM via the configured Sealer.
//
// SetKey must be called once at process boot before Value is used.
// Calling Value before SetKey returns an error.
func (e EncryptedString) Value() (driver.Value, error) {
	if !e.set {
		return nil, nil
	}
	cfg := loadConfig()
	if cfg == nil {
		return nil, errors.New("encryptedtype: SetKey not called")
	}
	return cfg.sealer.Seal([]byte(e.plain), nil)
}

// Scan implements [sql.Scanner]. NULL scans into an unset value. A
// non-NULL value is opened via crypt.OpenAuto so that both AES-256-GCM
// and AES-CBC ciphertexts decrypt transparently.
func (e *EncryptedString) Scan(src any) error {
	if src == nil {
		e.plain = ""
		e.set = false
		return nil
	}
	var ct string
	switch v := src.(type) {
	case string:
		ct = v
	case []byte:
		ct = string(v)
	default:
		return fmt.Errorf("encryptedtype: cannot scan %T", src)
	}
	if ct == "" {
		e.plain = ""
		e.set = false
		return nil
	}
	cfg := loadConfig()
	if cfg == nil {
		return errors.New("encryptedtype: SetKey not called")
	}
	pt, err := crypt.OpenAuto(cfg.key, ct, nil)
	if err != nil {
		return fmt.Errorf("encryptedtype: decrypt: %w", err)
	}
	e.plain = string(pt)
	e.set = true
	return nil
}

// String returns "[encrypted]" — the plaintext does not appear in fmt
// output, panic traces, or default Stringer paths.
func (e EncryptedString) String() string { return "[encrypted]" }

// GoString returns "[encrypted]".
func (e EncryptedString) GoString() string { return "[encrypted]" }

// MarshalJSON returns null. JSON consumers that want the plaintext
// must call Plain() and serialise it explicitly.
func (e EncryptedString) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

// UnmarshalJSON accepts a string (treated as plaintext) or null.
func (e *EncryptedString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		e.plain = ""
		e.set = false
		return nil
	}
	var s string
	if err := unmarshalString(data, &s); err != nil {
		return fmt.Errorf("encryptedtype: UnmarshalJSON: %w", err)
	}
	e.plain = s
	e.set = true
	return nil
}

// MarshalGQL writes the plaintext as a JSON string. The schema author
// opted into exposing the field via GraphQL; if the field should be
// server-only, mark it with the `@internal` directive in the schema.
func (e EncryptedString) MarshalGQL(w io.Writer) {
	if !e.set {
		_, _ = w.Write([]byte("null"))
		return
	}
	data, err := jsonString(e.plain)
	if err != nil {
		_, _ = w.Write([]byte("null"))
		return
	}
	_, _ = w.Write(data)
}

// UnmarshalGQL accepts a plaintext string from a GraphQL input field.
func (e *EncryptedString) UnmarshalGQL(v any) error {
	if v == nil {
		e.plain = ""
		e.set = false
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("encryptedtype: UnmarshalGQL: expected string, got %T", v)
	}
	e.plain = s
	e.set = true
	return nil
}

// --- module-level configuration ---

type config struct {
	sealer *crypt.Sealer
	key    []byte
}

var configRef atomic.Pointer[config]

// SetKey configures the AES key used by EncryptedString. The key must be
// exactly 16, 24, or 32 bytes — the same constraints crypt.NewSealer
// applies. Returns an error if the key is invalid.
//
// SetKey is goroutine-safe and may be called more than once (e.g. for
// key rotation); callers are responsible for ensuring no Value/Scan is
// in flight when the key changes.
func SetKey(key []byte) error {
	s, err := crypt.NewSealer(key)
	if err != nil {
		return fmt.Errorf("encryptedtype: SetKey: %w", err)
	}
	dup := make([]byte, len(key))
	copy(dup, key)
	configRef.Store(&config{sealer: s, key: dup})
	return nil
}

// SetSealer is a lower-level escape hatch for callers who already hold a
// *crypt.Sealer. Note that Scan uses crypt.OpenAuto which requires the
// raw key bytes; SetSealer alone disables the AES-CBC fallback path on
// reads. Prefer SetKey unless you have a specific reason.
func SetSealer(s *crypt.Sealer, key []byte) {
	dup := make([]byte, len(key))
	copy(dup, key)
	configRef.Store(&config{sealer: s, key: dup})
}

// Reset clears the stored key. Intended for tests that want to verify
// the "not configured" error path.
func Reset() { configRef.Store(nil) }

func loadConfig() *config { return configRef.Load() }
