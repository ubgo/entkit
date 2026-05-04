package entjsonslice

import (
	"testing"

	"entgo.io/ent/schema/field"
)

func TestField_Descriptor(t *testing.T) {
	f := Field("tags")
	d := f.Descriptor()
	if d.Name != "tags" {
		t.Fatalf("Name = %q, want tags", d.Name)
	}
	if d.Info == nil || d.Info.Type != field.TypeJSON {
		t.Fatalf("Info.Type = %v, want TypeJSON", d.Info)
	}
	if d.Err != nil {
		t.Fatalf("Descriptor.Err = %v", d.Err)
	}
}
