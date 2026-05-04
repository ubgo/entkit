# jsontype-ent

> Ent field helper for [`github.com/ubgo/jsontype`](https://github.com/ubgo/jsontype). One line of schema instead of remembering the right `field.JSON(name, jsontype.JSON{})` incantation.

## Install

```sh
go get github.com/ubgo/jsontype-ent
```

## Use

```go
import (
    "entgo.io/ent"
    "entgo.io/ent/schema"

    entjsontype "github.com/ubgo/jsontype-ent"
)

type Event struct{ ent.Schema }

func (Event) Fields() []ent.Field {
    return []ent.Field{
        entjsontype.Field("payload"),
    }
}
```

For finer control (Optional, Immutable, dialect-specific column type, comments, etc.) compose directly:

```go
field.JSON("payload", jsontype.JSON{}).
    Optional().
    Comment("free-form JSON event payload").
    SchemaType(map[string]string{"postgres": "jsonb"})
```

The helper is for the 90% case. The full ent API stays one import away.

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
