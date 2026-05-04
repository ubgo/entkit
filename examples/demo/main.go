// demo exercises every type in the family end-to-end without requiring
// generated ent code. Each section round-trips a value through the
// SQL Value/Scan boundary to prove it works in isolation.
package main

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/ubgo/entkit/encryptedtype"
	"github.com/ubgo/entkit/jsonmap"
	"github.com/ubgo/entkit/passwordtype"
	"github.com/ubgo/jsonslice"
	"github.com/ubgo/jsontype"
)

// scanValue simulates the database round-trip: take a driver.Value,
// turn it into the bytes the driver would deliver to Scan, and feed
// that to the type's Scan method.
func scanValue(s interface{ Scan(any) error }, v driver.Value) error {
	switch x := v.(type) {
	case nil:
		return s.Scan(nil)
	case string:
		return s.Scan(x)
	case []byte:
		return s.Scan(x)
	default:
		return fmt.Errorf("unsupported driver value type %T", v)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// 32 bytes — AES-256 key for encryptedtype.
	if err := encryptedtype.SetKey([]byte("0123456789abcdef0123456789abcdef")); err != nil {
		return fmt.Errorf("SetKey: %w", err)
	}

	if err := demoJSON(); err != nil {
		return fmt.Errorf("JSON: %w", err)
	}
	if err := demoJSONMap(); err != nil {
		return fmt.Errorf("JsonMap: %w", err)
	}
	if err := demoJSONSlice(); err != nil {
		return fmt.Errorf("JsonSlice: %w", err)
	}
	if err := demoHashedPassword(); err != nil {
		return fmt.Errorf("HashedPassword: %w", err)
	}
	if err := demoEncryptedString(); err != nil {
		return fmt.Errorf("EncryptedString: %w", err)
	}

	fmt.Println()
	fmt.Println("=== all five types round-tripped successfully ===")
	return nil
}

func demoJSON() error {
	src := jsontype.JSON(`{"role":"admin","plan":"pro"}`)
	val, err := src.Value()
	if err != nil {
		return err
	}
	var dst jsontype.JSON
	if err := scanValue(&dst, val); err != nil {
		return err
	}
	gqlOut := mustMarshalGQL(src)
	fmt.Printf("[jsontype.JSON]        Value=%v Scan=%s GQL=%s\n", val, string(dst), gqlOut)
	return nil
}

func demoJSONMap() error {
	src := jsonmap.JsonMap{"theme": "dark", "tier": float64(2)}
	val, err := src.Value()
	if err != nil {
		return err
	}
	var dst jsonmap.JsonMap
	if err := scanValue(&dst, val); err != nil {
		return err
	}
	out, _ := json.Marshal(dst)
	gqlOut := mustMarshalGQL(src)
	fmt.Printf("[jsonmap.JsonMap]      Value=%v Scan=%s GQL=%s\n", val, out, gqlOut)
	return nil
}

func demoJSONSlice() error {
	src := jsonslice.JsonSlice{"urgent", "billing", float64(42)}
	val, err := src.Value()
	if err != nil {
		return err
	}
	var dst jsonslice.JsonSlice
	if err := scanValue(&dst, val); err != nil {
		return err
	}
	out, _ := json.Marshal(dst)
	gqlOut := mustMarshalGQL(src)
	fmt.Printf("[jsonslice.JsonSlice]  Value=%v Scan=%s GQL=%s\n", val, out, gqlOut)
	return nil
}

func demoHashedPassword() error {
	src, err := passwordtype.New("hunter2")
	if err != nil {
		return err
	}
	val, err := src.Value()
	if err != nil {
		return err
	}
	var dst passwordtype.HashedPassword
	if err := scanValue(&dst, val); err != nil {
		return err
	}
	if !dst.Verify("hunter2") {
		return fmt.Errorf("Verify(correct) failed after round-trip")
	}
	if dst.Verify("wrong") {
		return fmt.Errorf("Verify(wrong) accepted")
	}
	gqlOut := mustMarshalGQL(src)
	fmt.Printf("[passwordtype]         Value=%T(len=%d) Verify(correct)=true Verify(wrong)=false GQL=%s String=%q\n",
		val, len(val.(string)), gqlOut, src.String())
	return nil
}

func demoEncryptedString() error {
	src := encryptedtype.New("rosebud-1234")
	val, err := src.Value()
	if err != nil {
		return err
	}
	var dst encryptedtype.EncryptedString
	if err := scanValue(&dst, val); err != nil {
		return err
	}
	if dst.Plain() != "rosebud-1234" {
		return fmt.Errorf("Plain() mismatch after round-trip: got %q", dst.Plain())
	}
	gqlOut := mustMarshalGQL(src)
	fmt.Printf("[encryptedtype]        Value=ciphertext(len=%d) Scan→Plain=%q GQL=%s String=%q\n",
		len(val.(string)), dst.Plain(), gqlOut, src.String())
	return nil
}

// mustMarshalGQL drives the duck-typed MarshalGQL and returns the
// resulting bytes. Each of our types implements
// MarshalGQL(io.Writer) — gqlgen detects this signature by name
// without us having to import gqlgen here.
func mustMarshalGQL(v interface{ MarshalGQL(io.Writer) }) string {
	var buf bytes.Buffer
	v.MarshalGQL(&buf)
	return buf.String()
}
