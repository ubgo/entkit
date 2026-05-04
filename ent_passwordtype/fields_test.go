package entpasswordtype

import (
	"testing"

	"entgo.io/ent/schema/field"
)

func TestField_Descriptor(t *testing.T) {
	f := Field("password")
	d := f.Descriptor()
	if d.Name != "password" {
		t.Fatalf("Name = %q, want password", d.Name)
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
