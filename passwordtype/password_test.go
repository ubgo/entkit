package passwordtype

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestHashedPassword_NewAndVerify(t *testing.T) {
	p, err := New("hunter2")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if !p.IsSet() {
		t.Fatal("IsSet() should be true after New")
	}
	if !p.Verify("hunter2") {
		t.Fatal("Verify should accept correct password")
	}
	if p.Verify("wrong") {
		t.Fatal("Verify must reject wrong password")
	}
}

func TestHashedPassword_New_RejectsEmpty(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("New(\"\") must error")
	}
}

func TestHashedPassword_ZeroValue(t *testing.T) {
	var p HashedPassword
	if p.IsSet() {
		t.Fatal("zero value must not be set")
	}
	if p.Verify("anything") {
		t.Fatal("zero value must reject any verify")
	}
	v, err := p.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if v != nil {
		t.Fatalf("zero Value() = %v, want nil", v)
	}
}

func TestHashedPassword_SQLRoundTrip(t *testing.T) {
	src, err := New("rosebud")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	v, err := src.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	stored, ok := v.(string)
	if !ok {
		t.Fatalf("Value type = %T, want string", v)
	}

	var dst HashedPassword
	if err := dst.Scan(stored); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !dst.Verify("rosebud") {
		t.Fatal("Scanned hash must verify against original plaintext")
	}
	if dst.Verify("notrosebud") {
		t.Fatal("Scanned hash must not verify against wrong plaintext")
	}

	var dst2 HashedPassword
	if err := dst2.Scan([]byte(stored)); err != nil {
		t.Fatalf("Scan []byte: %v", err)
	}
	if !dst2.Verify("rosebud") {
		t.Fatal("Scan []byte path failed verify")
	}
}

func TestHashedPassword_Scan_Errors(t *testing.T) {
	var p HashedPassword
	if err := p.Scan(42); err == nil {
		t.Fatal("Scan(int) must error")
	}
}

func TestHashedPassword_Scan_Nil(t *testing.T) {
	p, _ := New("set")
	if err := p.Scan(nil); err != nil {
		t.Fatalf("Scan(nil): %v", err)
	}
	if p.IsSet() {
		t.Fatal("Scan(nil) must clear the password")
	}
}

func TestHashedPassword_Redaction(t *testing.T) {
	p, _ := New("very-secret")

	if got := p.String(); got != "[redacted]" {
		t.Fatalf("String() = %q", got)
	}
	if got := p.GoString(); got != "[redacted]" {
		t.Fatalf("GoString() = %q", got)
	}

	jsonBytes, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(jsonBytes) != "null" {
		t.Fatalf("MarshalJSON = %s, want null", jsonBytes)
	}

	formatted := fmt.Sprintf("%v / %#v", p, p)
	if strings.Contains(formatted, "argon2") {
		t.Fatalf("fmt output leaks hash: %s", formatted)
	}
}

func TestHashedPassword_LogValue(t *testing.T) {
	p, _ := New("logtest")
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	logger.Info("user", "password", p)
	if strings.Contains(buf.String(), "argon2") {
		t.Fatalf("slog leaked hash: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "[redacted]") {
		t.Fatalf("slog did not redact: %s", buf.String())
	}
}

func TestHashedPassword_MarshalGQL(t *testing.T) {
	p, _ := New("anything")
	var buf bytes.Buffer
	p.MarshalGQL(&buf)
	if buf.String() != "null" {
		t.Fatalf("MarshalGQL = %q, want null", buf.String())
	}
}

func TestHashedPassword_UnmarshalGQL(t *testing.T) {
	var p HashedPassword
	if err := p.UnmarshalGQL("freshpass"); err != nil {
		t.Fatalf("UnmarshalGQL: %v", err)
	}
	if !p.IsSet() {
		t.Fatal("after UnmarshalGQL, password should be set")
	}
	if !p.Verify("freshpass") {
		t.Fatal("UnmarshalGQL did not hash plaintext correctly")
	}
	if p.Hash() == "freshpass" {
		t.Fatal("UnmarshalGQL stored plaintext instead of hashing")
	}
}

func TestHashedPassword_UnmarshalGQL_Errors(t *testing.T) {
	var p HashedPassword
	if err := p.UnmarshalGQL(42); err == nil {
		t.Fatal("UnmarshalGQL(int) must error")
	}
	if err := p.UnmarshalGQL(""); err == nil {
		t.Fatal("UnmarshalGQL(\"\") must error")
	}
}

func TestHashedPassword_UnmarshalGQL_Nil(t *testing.T) {
	p, _ := New("starting")
	if err := p.UnmarshalGQL(nil); err != nil {
		t.Fatalf("UnmarshalGQL(nil): %v", err)
	}
	if p.IsSet() {
		t.Fatal("UnmarshalGQL(nil) must clear the password")
	}
}

func TestHashedPassword_UnmarshalJSON(t *testing.T) {
	var p HashedPassword
	if err := p.UnmarshalJSON([]byte(`"jsonpass"`)); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if !p.Verify("jsonpass") {
		t.Fatal("UnmarshalJSON did not hash plaintext")
	}

	var p2 HashedPassword
	if err := p2.UnmarshalJSON([]byte(`null`)); err != nil {
		t.Fatalf("UnmarshalJSON(null): %v", err)
	}
	if p2.IsSet() {
		t.Fatal("null should leave password unset")
	}
}

func TestHashedPassword_FromHash(t *testing.T) {
	src, _ := New("orig")
	stored := src.Hash()

	wrapped := FromHash(stored)
	if !wrapped.Verify("orig") {
		t.Fatal("FromHash did not preserve verifiability")
	}
}
