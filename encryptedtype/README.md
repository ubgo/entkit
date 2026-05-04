# encryptedtype

> A SQL column type that transparently encrypts plaintext on write and decrypts on read. Doubles as a gqlgen scalar.

`EncryptedString` wraps [`github.com/ubgo/crypt`](https://github.com/ubgo/crypt). Writes use AES-256-GCM (`crypt.Sealer.Seal`). Reads use `crypt.OpenAuto`, which decrypts both AES-256-GCM (modern AEAD) and AES-CBC (peer format) so existing CBC-encrypted columns continue to read without migration.

| Channel | Behavior |
|---------|----------|
| `database/sql` `Value()` | Encrypts plaintext to AES-256-GCM ciphertext |
| `database/sql` `Scan()` | Decrypts via `crypt.OpenAuto` (AES-GCM **or** AES-CBC) |
| `encoding/json` `MarshalJSON` | Always `null` |
| `encoding/json` `UnmarshalJSON` | Accepts plaintext string |
| `fmt.Stringer` / `fmt.GoStringer` | `[encrypted]` |
| gqlgen `MarshalGQL` | Plaintext (schema author opted in) |
| gqlgen `UnmarshalGQL` | Plaintext input |
| `Plain() string` | Explicit accessor for in-memory plaintext |

Write path = AEAD. Read path = AEAD with CBC fallback. No migration step needed when adopting from a system that wrote CBC.

## When to use this vs `passwordtype`

| Use this | When |
|----------|------|
| `encryptedtype.EncryptedString` | The plaintext must be **recoverable** — API client secrets, OAuth refresh tokens, third-party API keys you'll send back out, anything you'd display in an admin UI |
| [`passwordtype.HashedPassword`](../passwordtype) | The plaintext only ever needs to be **verified**, never recovered — user authentication passwords |

If you're tempted to encrypt a user password, you want `passwordtype` instead.

## Install

```sh
go get github.com/ubgo/entkit/encryptedtype
```

Pair with [`ent_encryptedtype`](../ent_encryptedtype) when using Ent ORM.

## Boot wiring (required, exactly once)

The package needs an AES key. Wire it once at process startup:

```go
import "github.com/ubgo/entkit/encryptedtype"

func main() {
    key := []byte(os.Getenv("ENCRYPTION_KEY"))   // 32 bytes for AES-256
    if err := encryptedtype.SetKey(key); err != nil {
        log.Fatal("encryption key invalid: ", err)
    }
    // ... start server
}
```

The key must be 16, 24, or 32 bytes (AES-128 / AES-192 / AES-256). Calls to `Value()` or `Scan()` before `SetKey` return a clear error.

## Use case 1 — Storing API client secrets

A partner integrations table where each partner's API secret must be readable to forward to the partner's webhook endpoint.

```go
import "github.com/ubgo/entkit/encryptedtype"

type Partner struct {
    ID           string
    Name         string
    ClientSecret encryptedtype.EncryptedString
}

// Create
db.Exec(`INSERT INTO partners(id, name, client_secret) VALUES (?, ?, ?)`,
    p.ID, p.Name, encryptedtype.New(plaintextSecret))

// Read + use
var p Partner
db.QueryRow(`SELECT id, name, client_secret FROM partners WHERE id = ?`, id).
    Scan(&p.ID, &p.Name, &p.ClientSecret)
http.Post(webhookURL, withSignature(payload, p.ClientSecret.Plain()))
```

## Use case 2 — OAuth refresh tokens

```go
type OAuthAccount struct {
    UserID       string
    Provider     string
    RefreshToken encryptedtype.EncryptedString
    ExpiresAt    time.Time
}

// On token refresh from provider
acct.RefreshToken = encryptedtype.New(newRefreshToken)
db.Save(&acct)

// On next access-token request
fresh, err := provider.Exchange(ctx, acct.RefreshToken.Plain())
```

## Use case 3 — Reading legacy AES-CBC ciphertexts

If you have existing rows encrypted with AES-CBC (e.g. from `lace/crypt` or any other CBC-based scheme that uses the same key), they decrypt transparently on read:

```go
// row written years ago via crypt.EncryptCBC(key, plaintext)
var legacy encryptedtype.EncryptedString
db.QueryRow(`SELECT api_secret FROM old_partners WHERE id = ?`, id).Scan(&legacy)
fmt.Println(legacy.Plain())   // works — OpenAuto detected CBC and dispatched
```

New writes always use AES-256-GCM; the read path supports both. No big-bang migration step.

## Use case 4 — gqlgen field for admin UI

Schema:

```graphql
scalar EncryptedString

type Partner {
  id: ID!
  name: String!
  clientSecret: EncryptedString!   # plaintext returned to admin clients
}
```

`gqlgen.yml`:

```yaml
models:
  EncryptedString:
    model: github.com/ubgo/entkit/encryptedtype.EncryptedString
```

The output `clientSecret` field returns the plaintext via `MarshalGQL`. **Always** gate the resolver behind authn + authz so only operators see it. For server-only fields (never exposed to GraphQL), mark with the `@internal` directive in your schema and don't include the field in non-admin queries.

## Use case 5 — Rotating the encryption key

When you bump the encryption key (e.g. routine rotation), AES-CBC ciphertexts written with the old key continue to decrypt **only if the old and new keys match for the algorithm in question**. To genuinely rotate keys, walk every encrypted column and re-encrypt:

```go
func RotateKeys(ctx context.Context, oldKey, newKey []byte) error {
    encryptedtype.SetKey(oldKey)   // configure old key for read path

    rows, _ := db.Query(`SELECT id, client_secret FROM partners`)
    plaintexts := map[string]string{}
    for rows.Next() {
        var id string
        var v encryptedtype.EncryptedString
        rows.Scan(&id, &v)
        plaintexts[id] = v.Plain()
    }

    encryptedtype.SetKey(newKey)   // configure new key for write path
    for id, plain := range plaintexts {
        re := encryptedtype.New(plain)
        db.Exec(`UPDATE partners SET client_secret = ? WHERE id = ?`, re, id)
    }
    return nil
}
```

(Run during a maintenance window with all writes paused.)

## Use case 6 — In-memory secrets pipeline

`Plain()` exposes the in-memory plaintext to the same trusted call site that holds the value. Avoid copying the string out further; pass the `EncryptedString` itself across function boundaries when possible so the redaction surfaces (`String`, `MarshalJSON`) protect against accidental logging in transit.

```go
func chargeStripe(ctx context.Context, secret encryptedtype.EncryptedString, amount int64) error {
    client := stripe.NewClient(secret.Plain())   // one short-lived use
    return client.Charge(ctx, amount)
}
```

## API reference

| Function / method | Purpose |
|---|---|
| `SetKey(key []byte) error` | Configure AES key (16/24/32 bytes). Required before any read/write |
| `SetSealer(s *crypt.Sealer, key []byte)` | Lower-level escape hatch when you already hold a Sealer |
| `Reset()` | Clear stored key — for tests |
| `New(plaintext string) EncryptedString` | Wrap a plaintext value |
| `(e EncryptedString) Plain() string` | Read in-memory plaintext |
| `(e EncryptedString) IsSet() bool` | True when populated |
| `(e EncryptedString) Value() (driver.Value, error)` | DB write — encrypts to AES-256-GCM |
| `(e *EncryptedString) Scan(src any) error` | DB read — uses `crypt.OpenAuto` (AEAD or CBC) |
| `(e EncryptedString) MarshalJSON() ([]byte, error)` | Always `null` |
| `(e *EncryptedString) UnmarshalJSON([]byte) error` | Accepts plaintext string |
| `(e EncryptedString) String() string` | `[encrypted]` |
| `(e EncryptedString) GoString() string` | `[encrypted]` |
| `(e EncryptedString) MarshalGQL(io.Writer)` | Plaintext when set, `null` when unset |
| `(e *EncryptedString) UnmarshalGQL(any) error` | Plaintext input or null |

## Behaviour notes

- **Reads accept both AES-256-GCM and AES-CBC ciphertexts.** Driven by `crypt.OpenAuto`. No migration step needed.
- **Writes always use AES-256-GCM.** Future reads of new writes go through the AEAD path.
- **`MarshalGQL` returns plaintext.** This is intentional — exposing the field via GraphQL is a deliberate schema choice. Use the `@internal` directive in your schema for fields that should never be exposed.
- **Defense in depth on non-DB outputs.** `String()` returns `"[encrypted]"`, `MarshalJSON` returns `null`, `GoString()` returns `"[encrypted]"`. Plaintext only appears via `Plain()` (explicit) and `MarshalGQL` (schema opt-in).
- **Zero value is unset.** A `EncryptedString{}` returns `false` from `IsSet`, `nil` from `Value`, and `""` from `Plain`.
- **`Scan(nil)` clears** — useful for nullable columns.
- **Race-safe.** The internal config swap uses `atomic.Pointer`. Multiple goroutines running `Value`/`Scan` while a `SetKey` lands are safe (each operation observes either the old or the new config atomically).

## License

Apache-2.0 — see [`../LICENSE`](../LICENSE).
