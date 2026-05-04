package entjsontype

import (
	"testing"

	"entgo.io/ent/schema/field"
)

func TestField_Descriptor(t *testing.T) {
	f := Field("metadata")
	d := f.Descriptor()
	if d.Name != "metadata" {
		t.Fatalf("Name = %q, want metadata", d.Name)
	}
	if d.Info == nil || d.Info.Type != field.TypeJSON {
		t.Fatalf("Info.Type = %v, want TypeJSON", d.Info)
	}
	if d.Err != nil {
		t.Fatalf("Descriptor.Err = %v", d.Err)
	}
}
