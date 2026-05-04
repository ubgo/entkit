# ent_passwordtype

> Ent field helper for [`entkit/passwordtype`](../passwordtype). Returns a TEXT column tagged `Sensitive()` so ent's debug output never prints the argon2id hash.

## Install

```sh
go get github.com/ubgo/entkit/ent_passwordtype
```

## Use case 1 — User entity with login password

```go
import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"

    entpasswordtype "github.com/ubgo/entkit/ent_passwordtype"
)

type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("email").Unique(),
        entpasswordtype.Field("password"),
    }
}
```

After `task entg`, you get setters that take a `passwordtype.HashedPassword`:

```go
import "github.com/ubgo/entkit/passwordtype"

p, _ := passwordtype.New(input.Password)
client.User.Create().
    SetEmail(input.Email).
    SetPassword(p).
    Save(ctx)
```

## Use case 2 — Optional password (e.g. SSO-only users)

For schema variants where the password is optional, compose the field directly:

```go
import (
    "entgo.io/ent/schema/field"

    "github.com/ubgo/entkit/passwordtype"
)

field.String("password").
    GoType(passwordtype.HashedPassword{}).
    Sensitive().
    Optional()
```

## Use case 3 — Login flow (resolver)

```go
func (r *queryResolver) Login(ctx context.Context, email, password string) (*ent.User, error) {
    u, err := r.client.User.Query().Where(user.EmailEQ(email)).Only(ctx)
    if err != nil || !u.Password.Verify(password) {
        return nil, errInvalidCredentials   // intentionally same for both branches
    }
    return u, nil
}
```

## Use case 4 — gqlgen scalar (auto-hash on input, redact on output)

Schema:

```graphql
scalar HashedPassword

input SignupInput {
    email: String!
    password: HashedPassword!
}

type User {
    id: ID!
    email: String!
    password: HashedPassword     # always null in output
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

Resolver — note **zero** manual hashing or redaction:

```go
func (r *mutationResolver) Signup(ctx context.Context, input model.SignupInput) (*ent.User, error) {
    return r.client.User.Create().
        SetEmail(input.Email).
        SetPassword(input.Password).   // already a HashedPassword
        Save(ctx)
}
```

## Use case 5 — Multiple password fields (rotation period)

When you let users keep an old password active for grace-period rotations:

```go
func (User) Fields() []ent.Field {
    return []ent.Field{
        entpasswordtype.Field("password"),
        field.String("password_old").
            GoType(passwordtype.HashedPassword{}).
            Sensitive().
            Optional(),
        field.Time("password_old_expires_at").Optional(),
    }
}

// Login allows either while the old hasn't expired
if u.Password.Verify(p) || (u.PasswordOld.IsSet() &&
    !u.PasswordOldExpiresAt.IsZero() &&
    time.Now().Before(u.PasswordOldExpiresAt) &&
    u.PasswordOld.Verify(p)) {
    return u, nil
}
```

## API reference

| Function | Purpose |
|----------|---------|
| `Field(name string) ent.Field` | Returns an `ent.Field` for a `passwordtype.HashedPassword` column. Stored as TEXT. `Sensitive()` flag set. |

## License

Apache-2.0 — see [`../LICENSE`](../LICENSE).
