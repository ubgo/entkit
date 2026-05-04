# encryptedtype

> A SQL column type that transparently encrypts plaintext on write and decrypts on read. Doubles as a gqlgen scalar.

`EncryptedString` wraps [`github.com/ubgo/crypt`](https://github.com/ubgo/crypt). Writes use AES-256-GCM (`crypt.Sealer.Seal`). Reads use `crypt.OpenAuto`, which transparently decrypts both AES-256-GCM (modern AEAD) and AES-CBC (peer format) — so existing CBC-encrypted columns continue to read without migration.

| Integration | Hooks | Cost to your dependency tree |
|-------------|-------|------------------------------|
| `database/sql` | `Value()` / `Scan()` | stdlib + `ubgo/crypt` |
| `encoding/json` | `MarshalJSON` (always `null`) / `UnmarshalJSON` (accepts plaintext) | stdlib |
| gqlgen scalar | `MarshalGQL` (plaintext) / `UnmarshalGQL` (plaintext) | none — duck-typed |

## Install

```sh
go get github.com/ubgo/encryptedtype
```

## Boot wiring (required)

```go
import "github.com/ubgo/encryptedtype"

func main() {
    key := []byte(os.Getenv("ENCRYPTION_KEY"))   // 32 bytes for AES-256
    if err := encryptedtype.SetKey(key); err != nil {
        log.Fatal(err)
    }
    // ... start server
}
```

The key must be 16, 24, or 32 bytes (AES-128 / AES-192 / AES-256). New writes always use AES-256-GCM regardless of key length; the smaller key sizes still encrypt correctly, they just produce shorter sealing keys for the GCM cipher.

## Use

```go
import "github.com/ubgo/encryptedtype"

type Partner struct {
    ID           string
    ClientSecret encryptedtype.EncryptedString
}

// Save → ciphertext column
db.Exec(`INSERT INTO partners(id, client_secret) VALUES (?, ?)`,
    p.ID, encryptedtype.New(plaintext))

// Load → plaintext in memory
var loaded Partner
db.QueryRow(`SELECT id, client_secret FROM partners WHERE id = ?`, "x").
    Scan(&loaded.ID, &loaded.ClientSecret)
fmt.Println(loaded.ClientSecret.Plain())   // recovered plaintext
```

In a gqlgen `gqlgen.yml`:

```yaml
models:
  EncryptedString:
    model: github.com/ubgo/encryptedtype.EncryptedString
```

For server-only fields, mark with the `@internal` directive in your schema:

```graphql
scalar EncryptedString

type Partner {
  id: ID!
  clientSecret: EncryptedString! @internal
}
```

## Behaviour

- **Reads accept both AES-256-GCM and AES-CBC ciphertexts.** No migration step needed if you have existing CBC data.
- **Writes always use AES-256-GCM.** Future reads of new writes go through the AEAD path.
- **Defense in depth on non-DB outputs.** `String()` returns `"[encrypted]"`, `MarshalJSON` returns `null`, `GoString()` returns `"[encrypted]"`. The plaintext only appears via `Plain()` (explicit) and `MarshalGQL` (schema-author opt-in).
- **Zero value is unset.** A `EncryptedString{}` returns `false` from `IsSet`, `nil` from `Value`, and `""` from `Plain`.
- **`Scan(nil)` clears** — useful for nullable columns.

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
