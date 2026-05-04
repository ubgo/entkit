# encryptedtype-ent

> Ent field helper for [`github.com/ubgo/entkit/encryptedtype`](https://github.com/ubgo/entkit/encryptedtype).

Returns a TEXT column tagged `Sensitive()` so ent's debug output never prints the AES-GCM ciphertext.

## Install

```sh
go get github.com/ubgo/entkit/encryptedtype-ent
```

## Boot wiring (required)

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

## Use

```go
import (
    "entgo.io/ent"

    entencryptedtype "github.com/ubgo/entkit/encryptedtype-ent"
)

type Partner struct{ ent.Schema }

func (Partner) Fields() []ent.Field {
    return []ent.Field{
        entencryptedtype.Field("api_secret"),
    }
}
```

For finer control compose directly:

```go
field.String("api_secret").
    GoType(encryptedtype.EncryptedString{}).
    Sensitive().
    Optional()
```

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
