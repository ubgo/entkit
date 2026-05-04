// Package jsonsliceent provides an ent.Field constructor for
// github.com/ubgo/jsonslice.JsonSlice columns.
package entjsonslice

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/ubgo/jsonslice"
)

// Field returns an ent.Field for a jsonslice.JsonSlice column. The
// underlying ent type is Bytes (rather than JSON) so the generated
// code uses JsonSlice's Scanner/Valuer directly rather than re-running
// json.Unmarshal on top of an already-scanned slice. SchemaType pins
// the column to JSON / JSONB at the SQL layer for queryability.
//
// For Optional, Immutable, custom dialect column types, etc., compose
// directly:
//
//	field.Bytes("tags").GoType(jsonslice.JsonSlice{}).Optional().SchemaType(...)
func Field(name string) ent.Field {
	return field.Bytes(name).
		GoType(jsonslice.JsonSlice{}).
		SchemaType(map[string]string{
			"postgres": "jsonb",
			"mysql":    "json",
			"sqlite3":  "json",
		})
}
