// Package jsonmapent provides an ent.Field constructor for
// github.com/ubgo/entkit/jsonmap.JsonMap columns.
package entjsonmap

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/ubgo/entkit/jsonmap"
)

// Field returns an ent.Field for a jsonmap.JsonMap column. The column
// is stored as JSON / JSONB at the SQL layer.
func Field(name string) ent.Field {
	return field.JSON(name, jsonmap.JsonMap{})
}
