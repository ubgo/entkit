# passwordtype

> A SQL-friendly, gqlgen-friendly **argon2id** password type. Hashes plaintext on the way in, refuses to leak the hash on the way out.

`HashedPassword` wraps [`github.com/ubgo/crypt`](https://github.com/ubgo/crypt)'s argon2id `HashPassword` / `VerifyPassword`. Plaintext is hashed once, on entry, and is never recoverable.

| Integration | Hooks | Cost to your dependency tree |
|-------------|-------|------------------------------|
| `database/sql` | `Value()` / `Scan()` | stdlib + `ubgo/crypt` |
| `encoding/json` | `MarshalJSON` (always `null`) / `UnmarshalJSON` (hashes plaintext) | stdlib |
| `log/slog` | `LogValue()` returns `[redacted]` | stdlib |
| gqlgen scalar | `MarshalGQL` (always `null`) / `UnmarshalGQL` (hashes plaintext) | none — duck-typed |

## Install

```sh
go get github.com/ubgo/passwordtype
```

## Use

```go
import "github.com/ubgo/passwordtype"

// Signup
p, err := passwordtype.New(input.Password)
if err != nil { return err }
db.User.Create().SetPassword(p).Save(ctx)

// Login
user, _ := db.User.Query().Where(user.EmailEQ(email)).Only(ctx)
if !user.Password.Verify(input.Password) {
    return errUnauthorized
}
```

In a gqlgen `gqlgen.yml`:

```yaml
models:
  HashedPassword:
    model: github.com/ubgo/passwordtype.HashedPassword
```

In your GraphQL schema:

```graphql
scalar HashedPassword

input SignupInput {
  email: String!
  password: HashedPassword!
}

type User {
  id: ID!
  password: HashedPassword   # always null in output
}
```

The `UnmarshalGQL` hook hashes the plaintext as gqlgen parses the input. Your resolver receives a `HashedPassword` whose stored value is already the argon2id hash — no manual hashing in resolvers, no chance of accidentally logging plaintext.

## Defense in depth

Every channel that could leak the hash is closed:

| Path | Behavior |
|------|----------|
| `fmt.Sprint`, `fmt.Sprintf("%v", ...)` | `"[redacted]"` |
| `fmt.Sprintf("%#v", ...)` (GoString) | `"[redacted]"` |
| `encoding/json.Marshal` | `null` |
| `log/slog` | `"[redacted]"` |
| gqlgen output | `null` |

Outbound `Value()` (the SQL driver) returns the hash because that's the column's storage representation. There is no path back to plaintext.

## Behaviour

- **Empty plaintext rejected.** `New("")` returns an error rather than producing a hash of nothing.
- **Zero value is unset.** A `HashedPassword{}` returns `false` from `Verify`, `nil` from `Value`, and `false` from `IsSet`.
- **`Scan(nil)` clears** the password — useful for nullable columns.
- **`Hash()` exists** for trusted re-export only (admin tooling, snapshot import). Most callers should never use it.
- **`FromHash(stored)`** wraps an already-hashed string for cases where you read from a trusted source that bypasses `Scan`.

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
