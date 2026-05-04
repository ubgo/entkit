# ent_encryptedtype

> Ent field helper for [`entkit/encryptedtype`](../encryptedtype). Returns a TEXT column tagged `Sensitive()` so ent's debug output never prints the AES-GCM ciphertext.

## Install

```sh
go get github.com/ubgo/entkit/ent_encryptedtype
```

## Boot wiring (required — covered by `encryptedtype`)

Before any `Value`/`Scan` happens, configure the AES key once at process startup:

```go
import "github.com/ubgo/entkit/encryptedtype"

func main() {
    if err := encryptedtype.SetKey([]byte(os.Getenv("ENCRYPTION_KEY"))); err != nil {
        log.Fatal(err)
    }
    // ... start server
}
```

See [`encryptedtype`'s README](../encryptedtype) for full key-management details.

## Use case 1 — API client secret column

```go
import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"

    entencryptedtype "github.com/ubgo/entkit/ent_encryptedtype"
)

type Partner struct{ ent.Schema }

func (Partner) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").NotEmpty(),
        entencryptedtype.Field("api_secret"),
    }
}
```

```go
import "github.com/ubgo/entkit/encryptedtype"

client.Partner.Create().
    SetName("Acme Corp").
    SetAPISecret(encryptedtype.New(plaintextSecret)).
    Save(ctx)
```

## Use case 2 — OAuth refresh tokens

```go
type OAuthAccount struct{ ent.Schema }

func (OAuthAccount) Fields() []ent.Field {
    return []ent.Field{
        field.String("user_id"),
        field.String("provider"),
        entencryptedtype.Field("refresh_token"),
        field.Time("expires_at"),
    }
}

// Refresh path
acct.RefreshToken = encryptedtype.New(newRefreshToken)
client.OAuthAccount.UpdateOne(acct).
    SetRefreshToken(acct.RefreshToken).
    SetExpiresAt(time.Now().Add(time.Hour)).
    Save(ctx)
```

## Use case 3 — Optional / nullable

```go
import "entgo.io/ent/schema/field"
import "github.com/ubgo/entkit/encryptedtype"

field.String("api_secret").
    GoType(encryptedtype.EncryptedString{}).
    Sensitive().
    Optional()
```

## Use case 4 — Reading legacy AES-CBC ciphertexts (no migration step)

If you migrated from a system that wrote AES-CBC ciphertexts under the same key, they decrypt transparently — `encryptedtype.EncryptedString.Scan` uses `crypt.OpenAuto`, which detects format and dispatches.

```go
// Existing legacy_partners table populated by an older codebase using crypt.EncryptCBC
type LegacyPartner struct{ ent.Schema }

func (LegacyPartner) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").Unique(),
        entencryptedtype.Field("api_secret"),
    }
}

// Reading just works — the helper does not care that the column was
// written with CBC. Future writes (UpdateOne / Create) use AES-256-GCM.
p, _ := client.LegacyPartner.Query().Where(legacypartner.IDEQ(id)).Only(ctx)
fmt.Println(p.APISecret.Plain())
```

## Use case 5 — gqlgen exposed admin UI field

```graphql
scalar EncryptedString

type Partner {
  id: ID!
  name: String!
  apiSecret: EncryptedString!
}
```

```yaml
# gqlgen.yml
models:
  EncryptedString:
    model: github.com/ubgo/entkit/encryptedtype.EncryptedString
```

`MarshalGQL` returns plaintext on output — gate the resolver behind operator authn/authz, or split into a server-only `@internal`-marked field.

## Use case 6 — Server-only field via `@internal`

If the encrypted value should never be exposed via GraphQL, mark the field internal in your schema and don't expose it at all:

```graphql
directive @internal on FIELD_DEFINITION

type Partner {
  id: ID!
  name: String!
  apiSecret: EncryptedString! @internal   # blocked at the resolver gate
}
```

(Wire `@internal` enforcement in your resolver middleware so external tokens can't query the field.)

## API reference

| Function | Purpose |
|----------|---------|
| `Field(name string) ent.Field` | Returns an `ent.Field` for an `encryptedtype.EncryptedString` column. Stored as TEXT. `Sensitive()` flag set. |

## License

Apache-2.0 — see [`../LICENSE`](../LICENSE).
