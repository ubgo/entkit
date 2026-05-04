// Package schema shows how to use every ent helper from this family in
// a single User entity. It is a static example file — runnable code
// lives in cmd/demo.
package schema

import (
	"entgo.io/ent"

	entencryptedtype "github.com/ubgo/entkit/ent_encryptedtype"
	entjsonmap "github.com/ubgo/entkit/ent_jsonmap"
	entjsonslice "github.com/ubgo/entkit/ent_jsonslice"
	entjsontype "github.com/ubgo/entkit/ent_jsontype"
	entpasswordtype "github.com/ubgo/entkit/ent_passwordtype"
)

// User is an example entity that exercises every type in the family.
type User struct{ ent.Schema }

// Fields composes one helper from each repo.
func (User) Fields() []ent.Field {
	return []ent.Field{
		// JSON object — flat or nested
		entjsontype.Field("profile"),

		// JSON map — typed map[string]any
		entjsonmap.Field("metadata"),

		// JSON array — slice of arbitrary values
		entjsonslice.Field("tags"),

		// Argon2id-hashed password (one-way)
		entpasswordtype.Field("password"),

		// AES-256-GCM encrypted string (round-trips)
		entencryptedtype.Field("api_secret"),
	}
}
