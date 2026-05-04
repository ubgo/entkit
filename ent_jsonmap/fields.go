// Package jsonmapent provides an ent.Field constructor for
// github.com/ubgo/entkit/jsonmap.JsonMap columns.
package entjsonmap

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/ubgo/entkit/jsonmap"
)

// Field returns an ent.Field for a jsonmap.JsonMap column. The
// underlying ent type is Bytes (rather than JSON) so the generated
// code uses JsonMap's Scanner/Valuer directly rather than re-running
// json.Unmarshal on top of an already-scanned map. SchemaType pins
// the column to JSON / JSONB at the SQL layer for queryability.
func Field(name string) ent.Field {
	return field.Bytes(name).
		GoType(jsonmap.JsonMap{}).
		SchemaType(map[string]string{
			"postgres": "jsonb",
			"mysql":    "json",
			"sqlite3":  "json",
		})
}
