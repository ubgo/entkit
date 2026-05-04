// Package passwordtype provides HashedPassword — a one-way argon2id-hashed
// password column type for SQL databases that doubles as a gqlgen scalar.
//
// The type wraps github.com/ubgo/crypt's HashPassword/VerifyPassword
// (argon2id, OWASP-recommended modern password hash). Plaintext is never
// recoverable from a HashedPassword; the only operations on the stored
// value are Verify and round-trip through SQL.
//
// Defense in depth: every redaction path is closed.
//
//   - String, GoString return "[redacted]"
//   - MarshalJSON returns null
//   - LogValue returns "[redacted]" so slog never emits the hash
//   - MarshalGQL writes null so the hash never reaches a GraphQL response
//
// On the input side, UnmarshalGQL accepts a plaintext string from a
// GraphQL input field, hashes it via crypt.HashPassword, and stores the
// hash. Resolvers therefore receive a HashedPassword that is already
// hashed — no manual hashing in business code, no temptation to log
// the plaintext.
//
// gqlgen integration is duck-typed: this package does not import gqlgen.
package passwordtype

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/ubgo/crypt"
)

// HashedPassword is an argon2id password hash. The plaintext is never
// stored. The zero value is unset (Value returns NULL, Verify returns
// false).
type HashedPassword struct {
	hash string
}

// New hashes plaintext with argon2id and returns a HashedPassword.
func New(plaintext string) (HashedPassword, error) {
	if plaintext == "" {
		return HashedPassword{}, errors.New("passwordtype: plaintext must not be empty")
	}
	h, err := crypt.HashPassword(plaintext)
	if err != nil {
		return HashedPassword{}, fmt.Errorf("passwordtype: hash: %w", err)
	}
	return HashedPassword{hash: h}, nil
}

// FromHash wraps an already-hashed PHC string. Use when reading from
// trusted storage that bypasses Scan (e.g. a snapshot import).
func FromHash(stored string) HashedPassword {
	return HashedPassword{hash: stored}
}

// Verify reports whether plaintext matches the stored hash. Returns
// false when the password is unset or any comparison error occurs;
// callers who need to distinguish the two cases should call IsSet
// first.
func (p HashedPassword) Verify(plaintext string) bool {
	if p.hash == "" {
		return false
	}
	ok, _ := crypt.VerifyPassword(plaintext, p.hash)
	return ok
}

// IsSet reports whether the password has been set.
func (p HashedPassword) IsSet() bool { return p.hash != "" }

// Hash returns the raw stored hash. Intended only for callers that must
// re-export the value via a trusted boundary (e.g. user-export tooling
// behind admin auth). Most callers should prefer Verify.
func (p HashedPassword) Hash() string { return p.hash }

// Value implements [driver.Valuer]. An unset password persists as SQL
// NULL.
func (p HashedPassword) Value() (driver.Value, error) {
	if p.hash == "" {
		return nil, nil
	}
	return p.hash, nil
}

// Scan implements [sql.Scanner]. Accepts string, []byte, or nil.
func (p *HashedPassword) Scan(src any) error {
	if src == nil {
		p.hash = ""
		return nil
	}
	switch v := src.(type) {
	case string:
		p.hash = v
	case []byte:
		p.hash = string(v)
	default:
		return fmt.Errorf("passwordtype: cannot scan %T into HashedPassword", src)
	}
	return nil
}

// String returns "[redacted]" — the stored hash never appears in fmt
// output, debug prints, or panic traces.
func (p HashedPassword) String() string { return "[redacted]" }

// GoString returns "[redacted]" — same protection for %#v formatting.
func (p HashedPassword) GoString() string { return "[redacted]" }

// MarshalJSON always returns null. There is no legitimate JSON consumer
// that needs the raw hash; encode the entity without the password
// field, or expose Verify behind an authorization check.
func (p HashedPassword) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

// UnmarshalJSON accepts either null or a string. A string is treated as
// a plaintext password and hashed.
func (p *HashedPassword) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		p.hash = ""
		return nil
	}
	var plain string
	if err := unmarshalString(data, &plain); err != nil {
		return fmt.Errorf("passwordtype: UnmarshalJSON: %w", err)
	}
	hp, err := New(plain)
	if err != nil {
		return err
	}
	*p = hp
	return nil
}

// LogValue makes HashedPassword safe for use with log/slog. The slog
// handler sees only "[redacted]", never the hash.
func (p HashedPassword) LogValue() slog.Value {
	return slog.StringValue("[redacted]")
}

// MarshalGQL is the gqlgen-compatible marshal hook. It always writes
// null — the hash must never leave the server via GraphQL.
func (p HashedPassword) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte("null"))
}

// UnmarshalGQL accepts a plaintext string from a GraphQL input field
// and hashes it. Resolvers receive a HashedPassword whose Value is the
// hash, not the plaintext.
func (p *HashedPassword) UnmarshalGQL(v any) error {
	if v == nil {
		p.hash = ""
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("passwordtype: UnmarshalGQL: expected string, got %T", v)
	}
	hp, err := New(s)
	if err != nil {
		return err
	}
	*p = hp
	return nil
}
