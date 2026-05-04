// Package schema is the ent schema for the entkit examples app. It
// composes one helper from every entkit ent_* sub-module to demonstrate
// the full type family in a single User entity.
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	entencryptedtype "github.com/ubgo/entkit/ent_encryptedtype"
	entjsonmap "github.com/ubgo/entkit/ent_jsonmap"
	entjsonslice "github.com/ubgo/entkit/ent_jsonslice"
	entjsontype "github.com/ubgo/entkit/ent_jsontype"
	entpasswordtype "github.com/ubgo/entkit/ent_passwordtype"
)

// User exercises every column type in the entkit family.
type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").Unique(),

		// JSON object whose keys vary per row (entkit/jsonmap).
		entjsonmap.Field("metadata"),

		// JSON array (ubgo/jsonslice).
		entjsonslice.Field("tags"),

		// Opaque JSON blob (ubgo/jsontype).
		entjsontype.Field("profile"),

		// Argon2id-hashed password — one-way, redacted on output.
		entpasswordtype.Field("password"),

		// AES-256-GCM encrypted string — recoverable plaintext.
		entencryptedtype.Field("api_secret"),
	}
}
