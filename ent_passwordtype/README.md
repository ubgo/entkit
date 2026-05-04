# passwordtype-ent

> Ent field helper for [`github.com/ubgo/entkit/passwordtype`](https://github.com/ubgo/entkit/passwordtype).

Returns a TEXT column tagged `Sensitive()` so ent's debug output never prints the argon2id hash.

## Install

```sh
go get github.com/ubgo/entkit/passwordtype-ent
```

## Use

```go
import (
    "entgo.io/ent"

    entpasswordtype "github.com/ubgo/entkit/passwordtype-ent"
)

type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
    return []ent.Field{
        entpasswordtype.Field("password"),
    }
}
```

For finer control compose directly:

```go
field.String("password").
    GoType(passwordtype.HashedPassword{}).
    Sensitive().
    Optional()
```

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
