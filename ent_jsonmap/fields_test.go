package entjsonmap

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
	if d.Info == nil || d.Info.Type != field.TypeBytes {
		t.Fatalf("Info.Type = %v, want TypeBytes", d.Info)
	}
	if d.SchemaType["postgres"] != "jsonb" {
		t.Fatalf("postgres SchemaType = %q, want jsonb", d.SchemaType["postgres"])
	}
	if d.Err != nil {
		t.Fatalf("Descriptor.Err = %v", d.Err)
	}
}
