package entencryptedtype

import (
	"testing"

	"entgo.io/ent/schema/field"

	"github.com/ubgo/entkit/encryptedtype"
)

func init() {
	// SetKey is required before any Scan/Value path. The schema-side
	// helpers do not call Scan/Value themselves but the GoType()
	// validation may walk type methods, so wire a key for safety.
	_ = encryptedtype.SetKey([]byte("0123456789abcdef0123456789abcdef"))
}

func TestField_Descriptor(t *testing.T) {
	f := Field("api_secret")
	d := f.Descriptor()
	if d.Name != "api_secret" {
		t.Fatalf("Name = %q, want api_secret", d.Name)
	}
	if d.Info == nil || d.Info.Type != field.TypeString {
		t.Fatalf("Info.Type = %v, want TypeString", d.Info)
	}
	if !d.Sensitive {
		t.Fatal("Sensitive flag must be set")
	}
	if d.Err != nil {
		t.Fatalf("Descriptor.Err = %v", d.Err)
	}
}
