// Package passwordtypeent provides an ent.Field constructor for
// github.com/ubgo/entkit/passwordtype.HashedPassword columns.
package entpasswordtype

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/ubgo/entkit/passwordtype"
)

// Field returns an ent.Field for a HashedPassword column.
//
//   - Stored as TEXT (the argon2id PHC string is variable-length).
//   - Sensitive() set so the value is masked in ent's debug/print output.
//
// For Optional, Immutable, dialect-specific column type, etc., compose
// directly:
//
//	field.String("password").
//	    GoType(passwordtype.HashedPassword{}).
//	    Sensitive().
//	    Optional()
func Field(name string) ent.Field {
	return field.String(name).
		GoType(passwordtype.HashedPassword{}).
		Sensitive()
}
