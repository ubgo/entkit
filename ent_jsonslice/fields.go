// Package jsonsliceent provides an ent.Field constructor for
// github.com/ubgo/jsonslice.JsonSlice columns.
package entjsonslice

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/ubgo/jsonslice"
)

// Field returns an ent.Field for a jsonslice.JsonSlice column. The
// column is stored as JSON / JSONB at the SQL layer.
//
// For Optional, Immutable, dialect-specific column type, etc., compose
// directly:
//
//	field.JSON("tags", jsonslice.JsonSlice{}).Optional().SchemaType(...)
func Field(name string) ent.Field {
	return field.JSON(name, jsonslice.JsonSlice{})
}
