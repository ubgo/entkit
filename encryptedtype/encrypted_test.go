package encryptedtype

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ubgo/crypt"
)

// 32-byte test key — fixed so test fixtures are reproducible.
var testKey = []byte("0123456789abcdef0123456789abcdef")

func init() {
	if err := SetKey(testKey); err != nil {
		panic(err)
	}
}

func TestEncryptedString_RoundTrip(t *testing.T) {
	src := New("hunter2")
	v, err := src.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	ct, ok := v.(string)
	if !ok {
		t.Fatalf("Value type = %T, want string", v)
	}
	if ct == "hunter2" {
		t.Fatal("Value returned plaintext")
	}

	var dst EncryptedString
	if err := dst.Scan(ct); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if !dst.IsSet() {
		t.Fatal("Scan should set IsSet")
	}
	if dst.Plain() != "hunter2" {
		t.Fatalf("Plain() = %q, want hunter2", dst.Plain())
	}
}

func TestEncryptedString_Scan_BytesAndString(t *testing.T) {
	src := New("payload")
	v, _ := src.Value()
	ct := v.(string)

	var d1 EncryptedString
	if err := d1.Scan(ct); err != nil {
		t.Fatalf("Scan(string): %v", err)
	}
	if d1.Plain() != "payload" {
		t.Fatalf("Plain after string scan: %q", d1.Plain())
	}

	var d2 EncryptedString
	if err := d2.Scan([]byte(ct)); err != nil {
		t.Fatalf("Scan([]byte): %v", err)
	}
	if d2.Plain() != "payload" {
		t.Fatalf("Plain after []byte scan: %q", d2.Plain())
	}
}

func TestEncryptedString_Scan_Nil(t *testing.T) {
	v := New("set")
	if err := v.Scan(nil); err != nil {
		t.Fatalf("Scan(nil): %v", err)
	}
	if v.IsSet() {
		t.Fatal("Scan(nil) must clear")
	}
	if v.Plain() != "" {
		t.Fatalf("Plain after nil = %q", v.Plain())
	}
}

func TestEncryptedString_Scan_EmptyString(t *testing.T) {
	var v EncryptedString
	if err := v.Scan(""); err != nil {
		t.Fatalf("Scan(\"\"): %v", err)
	}
	if v.IsSet() {
		t.Fatal("empty string should clear")
	}
}

func TestEncryptedString_Scan_UnsupportedType(t *testing.T) {
	var v EncryptedString
	if err := v.Scan(42); err == nil {
		t.Fatal("Scan(int) must error")
	}
}

func TestEncryptedString_Scan_BadCiphertext(t *testing.T) {
	var v EncryptedString
	if err := v.Scan("not-actually-ciphertext"); err == nil {
		t.Fatal("Scan(garbage) must error")
	}
}

func TestEncryptedString_Scan_AcceptsCBCCiphertext(t *testing.T) {
	cbcCT, err := crypt.EncryptCBC(testKey, []byte("legacy-secret"))
	if err != nil {
		t.Fatalf("EncryptCBC: %v", err)
	}

	var v EncryptedString
	if err := v.Scan(cbcCT); err != nil {
		t.Fatalf("Scan(CBC): %v", err)
	}
	if v.Plain() != "legacy-secret" {
		t.Fatalf("Plain = %q after CBC scan, want legacy-secret", v.Plain())
	}
}

func TestEncryptedString_ZeroValue_Value(t *testing.T) {
	var v EncryptedString
	val, err := v.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if val != nil {
		t.Fatalf("zero Value() = %v, want nil", val)
	}
}

func TestEncryptedString_NotConfigured(t *testing.T) {
	defer SetKey(testKey)
	Reset()

	v := New("x")
	if _, err := v.Value(); err == nil {
		t.Fatal("Value should error when SetKey not called")
	}
	var d EncryptedString
	if err := d.Scan("anything"); err == nil {
		t.Fatal("Scan should error when SetKey not called")
	}
}

func TestEncryptedString_SetKey_Invalid(t *testing.T) {
	if err := SetKey([]byte("too-short")); err == nil {
		t.Fatal("SetKey with short key must error")
	}
}

func TestEncryptedString_Redaction(t *testing.T) {
	v := New("very-secret")

	if v.String() != "[encrypted]" {
		t.Fatalf("String = %q", v.String())
	}
	if v.GoString() != "[encrypted]" {
		t.Fatalf("GoString = %q", v.GoString())
	}

	jsonBytes, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if string(jsonBytes) != "null" {
		t.Fatalf("MarshalJSON = %s", jsonBytes)
	}
}

func TestEncryptedString_MarshalGQL(t *testing.T) {
	var buf bytes.Buffer
	New("plaintext-out").MarshalGQL(&buf)
	if buf.String() != `"plaintext-out"` {
		t.Fatalf("MarshalGQL = %q", buf.String())
	}

	var unset EncryptedString
	buf.Reset()
	unset.MarshalGQL(&buf)
	if buf.String() != "null" {
		t.Fatalf("unset MarshalGQL = %q", buf.String())
	}
}

func TestEncryptedString_UnmarshalGQL(t *testing.T) {
	var v EncryptedString
	if err := v.UnmarshalGQL("plaintext-in"); err != nil {
		t.Fatalf("UnmarshalGQL: %v", err)
	}
	if v.Plain() != "plaintext-in" {
		t.Fatalf("Plain = %q", v.Plain())
	}

	var v2 EncryptedString
	if err := v2.UnmarshalGQL(nil); err != nil {
		t.Fatalf("UnmarshalGQL(nil): %v", err)
	}
	if v2.IsSet() {
		t.Fatal("nil should clear")
	}

	var v3 EncryptedString
	if err := v3.UnmarshalGQL(42); err == nil {
		t.Fatal("UnmarshalGQL(int) must error")
	}
}

func TestEncryptedString_UnmarshalJSON(t *testing.T) {
	var v EncryptedString
	if err := v.UnmarshalJSON([]byte(`"jsonpt"`)); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if v.Plain() != "jsonpt" {
		t.Fatalf("Plain after JSON unmarshal = %q", v.Plain())
	}

	if err := v.UnmarshalJSON([]byte(`null`)); err != nil {
		t.Fatalf("UnmarshalJSON(null): %v", err)
	}
	if v.IsSet() {
		t.Fatal("null should clear")
	}
}

func TestEncryptedString_NoFmtLeak(t *testing.T) {
	v := New("topsecret")
	got := strings.Join([]string{
		v.String(),
		v.GoString(),
	}, " | ")
	if strings.Contains(got, "topsecret") {
		t.Fatalf("plaintext leaked via String/GoString: %s", got)
	}
}
