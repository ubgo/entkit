// Package jsontypeent provides an ent.Field constructor for
// github.com/ubgo/jsontype.JSON columns.
//
// The helper is intentionally narrow: it returns a sensibly-defaulted
// field. For finer control (Optional, Immutable, Comment, StorageKey,
// SchemaType per dialect, etc.) compose directly:
//
//	field.JSON("metadata", jsontype.JSON{}).Optional().Comment("...")
package entjsontype

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/ubgo/jsontype"
)

// Field returns an ent.Field for a jsontype.JSON column. The column is
// stored as JSON / JSONB at the SQL layer; ent picks the appropriate
// type for the dialect.
func Field(name string) ent.Field {
	return field.JSON(name, jsontype.JSON{})
}
