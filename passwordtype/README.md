# passwordtype

> A SQL-friendly, gqlgen-friendly **argon2id** password type. Hashes plaintext on the way in, refuses to leak the hash on the way out.

`HashedPassword` wraps [`github.com/ubgo/crypt`](https://github.com/ubgo/crypt)'s argon2id `HashPassword` / `VerifyPassword`. Plaintext is hashed once, on entry, and is never recoverable.

| Channel | Behavior |
|---------|----------|
| `database/sql` `Value()` | Stores the argon2id PHC string |
| `database/sql` `Scan()` | Reads the stored hash |
| `encoding/json` `MarshalJSON` | Always `null` |
| `encoding/json` `UnmarshalJSON` | Hashes plaintext input |
| `log/slog` `LogValue` | Always `[redacted]` |
| `fmt.Stringer` / `fmt.GoStringer` | Always `[redacted]` |
| gqlgen `MarshalGQL` | Always `null` |
| gqlgen `UnmarshalGQL` | Hashes plaintext input |
| `Verify(plain) bool` | The only legitimate way to compare a candidate against the stored hash |

The `Hash()` accessor exists for trusted boundary use (admin export, snapshot import) but most callers should never reach for it.

## Install

```sh
go get github.com/ubgo/entkit/passwordtype
```

Pair with [`ent_passwordtype`](../ent_passwordtype) when using Ent ORM.

## Use case 1 — User signup

```go
import "github.com/ubgo/entkit/passwordtype"

func Signup(ctx context.Context, input SignupInput) (*User, error) {
    p, err := passwordtype.New(input.Password)
    if err != nil {
        return nil, fmt.Errorf("hash password: %w", err)
    }
    return db.User.Create().
        SetEmail(input.Email).
        SetPassword(p).
        Save(ctx)
}
```

## Use case 2 — Login

```go
func Login(ctx context.Context, email, plain string) (*User, error) {
    u, err := db.User.Query().Where(user.EmailEQ(email)).Only(ctx)
    if err != nil {
        return nil, errInvalidCredentials   // intentionally same error as wrong password
    }
    if !u.Password.Verify(plain) {
        return nil, errInvalidCredentials
    }
    return u, nil
}
```

`Verify` runs in constant time (delegated to `crypt.VerifyPassword`'s argon2 comparison).

## Use case 3 — Password change

```go
func ChangePassword(ctx context.Context, userID, oldPlain, newPlain string) error {
    u, err := db.User.Get(ctx, userID)
    if err != nil { return err }

    if !u.Password.Verify(oldPlain) {
        return errWrongOldPassword
    }
    newHash, err := passwordtype.New(newPlain)
    if err != nil { return err }

    return db.User.UpdateOne(u).SetPassword(newHash).Exec(ctx)
}
```

## Use case 4 — gqlgen mutation (zero resolver code for hashing)

GraphQL schema:

```graphql
scalar HashedPassword

input SignupInput {
    email: String!
    password: HashedPassword!
}

type User {
    id: ID!
    email: String!
    password: HashedPassword     # always null in output — defense in depth
}

type Mutation {
    signup(input: SignupInput!): User!
}
```

`gqlgen.yml`:

```yaml
models:
  HashedPassword:
    model: github.com/ubgo/entkit/passwordtype.HashedPassword
```

Resolver:

```go
func (r *mutationResolver) Signup(ctx context.Context, input model.SignupInput) (*ent.User, error) {
    return r.client.User.Create().
        SetEmail(input.Email).
        SetPassword(input.Password).   // already a HashedPassword — UnmarshalGQL hashed it
        Save(ctx)
}
```

The `password` field on the output type marshals as `null` even though ent's `User.Password` holds the argon2id hash — defense in depth for accidental queries.

## Use case 5 — Logging without leaking the hash

```go
slog.Info("user authenticated",
    slog.String("user_id", u.ID),
    slog.Any("password", u.Password),   // emits "[redacted]", never the hash
)
```

`LogValue()` is implemented; `slog`'s `Any` walks it before any custom handler sees the value.

## Use case 6 — Admin tooling that must export the hash

For migration / backup tooling that runs behind a trusted auth boundary:

```go
hash := u.Password.Hash()
exportRow := []string{u.ID, u.Email, hash}
```

`Hash()` is the one explicit accessor that returns the raw PHC string. Code review should treat any new caller of `Hash()` as a security-sensitive change.

## Use case 7 — Importing legacy hashes

If you're migrating from an existing system that already stores argon2id PHC strings:

```go
existing := readLegacyRow(...)
p := passwordtype.FromHash(existing.HashedPassword)

// Now Verify works against the legacy hash
if p.Verify(candidate) { ... }
```

For bcrypt or other algorithms, hash a fresh value with `passwordtype.New(plain)` on next successful login.

## API reference

| Function / method | Purpose |
|---|---|
| `New(plaintext string) (HashedPassword, error)` | Hash plaintext via argon2id (rejects empty input) |
| `FromHash(stored string) HashedPassword` | Wrap an already-hashed PHC string from trusted source |
| `(p HashedPassword) Verify(plaintext string) bool` | Compare candidate plaintext against stored hash |
| `(p HashedPassword) IsSet() bool` | True if hash has been set |
| `(p HashedPassword) Hash() string` | Trusted-boundary accessor — returns the raw PHC string |
| `(p HashedPassword) Value() (driver.Value, error)` | DB write — stores hash, NULL when unset |
| `(p *HashedPassword) Scan(src any) error` | DB read — accepts string, []byte, nil |
| `(p HashedPassword) MarshalJSON() ([]byte, error)` | Always `null` |
| `(p *HashedPassword) UnmarshalJSON([]byte) error` | Hashes plaintext; null → unset |
| `(p HashedPassword) String() string` | `[redacted]` |
| `(p HashedPassword) GoString() string` | `[redacted]` |
| `(p HashedPassword) LogValue() slog.Value` | `[redacted]` |
| `(p HashedPassword) MarshalGQL(io.Writer)` | Always `null` |
| `(p *HashedPassword) UnmarshalGQL(any) error` | Hashes plaintext; nil → unset |

## Behaviour notes

- **Empty plaintext rejected.** `New("")` returns an error rather than producing a hash of an empty string.
- **Zero value is unset.** A `HashedPassword{}` returns `false` from `Verify`, `nil` from `Value`, and `false` from `IsSet`.
- **`Scan(nil)` clears** the password — useful for nullable columns.
- **Verify is timing-safe.** Argon2id's verify path uses constant-time comparison.
- **Argon2id parameters are baked in via `ubgo/crypt`** (currently 64 MiB memory, 2 iterations, 1 lane). The PHC string self-describes parameters so future tuning doesn't break old hashes.

## Why argon2id, not bcrypt?

Argon2id is the OWASP-recommended modern password hash function. Bcrypt is acceptable for legacy compatibility but has a 72-byte input cap and weaker memory-hardness. If you're migrating from bcrypt, hash a fresh argon2id value on the user's next successful login.

## License

Apache-2.0 — see [`../LICENSE`](../LICENSE).
