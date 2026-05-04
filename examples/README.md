# entkit examples

> Runnable end-to-end demo. Opens an in-memory SQLite database, generates ent code, runs the schema migration, creates a `User` row exercising every entkit column type, reads it back through ent, and verifies every round-trip.

## Quick start

```sh
# 1. Generate the ent client (one-time per schema change)
task generate

# 2. Run the demo (auto-migrates + creates + reads + verifies)
task demo
```

Or directly:

```sh
go generate ./ent
go run ./cmd/demo
```

Sample output:

```
✓ schema migrated
✓ created user id=1 email=alice@example.com
✓ loaded user id=1
  ✓ email
  ✓ metadata.tier == 2
  ✓ tags[0] == admin
  ✓ profile contains bio
  ✓ password verifies (hunter2)
  ✓ wrong password rejected
  ✓ password.String() redacted
  ✓ password JSON is null
  ✓ api_secret decrypted
  ✓ api_secret.String() redacted
  ✓ api_secret JSON is null
  ✓ new password verifies (rosebud)
  ✓ old password rejected after change

=== all checkpoints passed — entkit family works end-to-end through ent + SQLite ===
```

## What the demo proves

The demo is one program that walks every layer of the entkit family:

1. **Boot wiring.** `encryptedtype.SetKey` configures the AES key once.
2. **Schema migration.** `client.Schema.Create` runs ent's auto-migration against in-memory SQLite — no external DB needed.
3. **Insert.** `client.User.Create()` builds a row with every column type populated:
   - `metadata`: `jsonmap.JsonMap` (dynamic JSON object)
   - `tags`: `jsonslice.JsonSlice` (dynamic JSON array, mixed types)
   - `profile`: `jsontype.JSON` (raw JSON blob)
   - `password`: `passwordtype.HashedPassword` (argon2id one-way hash)
   - `api_secret`: `encryptedtype.EncryptedString` (AES-256-GCM)
4. **Query + reconstitution.** `client.User.Query().Only(ctx)` round-trips every column through `Scan`. The hashed password and encrypted string each rebuild correctly.
5. **Behavioural assertions.** Each column-type assertion runs as its own checkpoint:
   - `Verify(plain)` accepts the original password, rejects a wrong one.
   - `Plain()` recovers the original encrypted plaintext.
   - `String()` and `MarshalJSON` redact sensitive types.
6. **Update path.** `client.User.UpdateOne().SetPassword(...)` rewrites the password column. The new hash verifies; the old plaintext no longer does.

If anything in the chain breaks, the demo exits with a non-zero status and prints the specific checkpoint that failed.

## Repository layout

```
examples/
├── go.mod              ← own module
├── Taskfile.yml        ← `task generate`, `task demo`, `task ci`
│
├── ent/
│   ├── schema/
│   │   └── user.go     ← single ent.Schema with all five column types
│   ├── generate.go     ← go:generate directive driving `ent generate`
│   ├── client.go       ← (generated) ent client
│   ├── user.go         ← (generated) User entity + scan/assign
│   └── ...             ← (generated) migrations, predicates, hooks, etc.
│
└── cmd/
    └── demo/
        └── main.go     ← runnable end-to-end demo
```

## Schema definition

`ent/schema/user.go` composes one helper from every entkit `ent_*` sub-module:

```go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"

    entencryptedtype "github.com/ubgo/entkit/ent_encryptedtype"
    entjsonmap       "github.com/ubgo/entkit/ent_jsonmap"
    entjsonslice     "github.com/ubgo/entkit/ent_jsonslice"
    entjsontype      "github.com/ubgo/entkit/ent_jsontype"
    entpasswordtype  "github.com/ubgo/entkit/ent_passwordtype"
)

type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("email").Unique(),
        entjsonmap.Field("metadata"),
        entjsonslice.Field("tags"),
        entjsontype.Field("profile"),
        entpasswordtype.Field("password"),
        entencryptedtype.Field("api_secret"),
    }
}
```

After `task generate`, ent emits the typed client API (`SetMetadata`, `SetPassword`, etc.) that the demo program drives.

## Adapting to your project

Copy the schema file into your own ent project, swap SQLite for Postgres / MySQL in `ent.Open`, and adjust the entity to match your domain. The generated column types (`jsonb` on Postgres, `json` on MySQL, etc.) are handled by each helper's `SchemaType` map.

## License

Apache-2.0 — see [`../LICENSE`](../LICENSE).
