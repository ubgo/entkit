# ent_jsonmap

> Ent field helper for [`entkit/jsonmap`](../jsonmap). For JSON **object** columns where keys vary per row.

## Install

```sh
go get github.com/ubgo/entkit/ent_jsonmap
```

## Use case 1 — User profile metadata

```go
import (
    "entgo.io/ent"

    entjsonmap "github.com/ubgo/entkit/ent_jsonmap"
)

type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
    return []ent.Field{
        entjsonmap.Field("metadata"),
    }
}
```

```go
client.User.Create().
    SetEmail(input.Email).
    SetMetadata(jsonmap.JsonMap{
        "stripe_id":     "cus_abc",
        "feature_flags": map[string]any{"new_dashboard": true},
    }).
    Save(ctx)
```

## Use case 2 — Webhook payload

```go
type WebhookEvent struct{ ent.Schema }

func (WebhookEvent) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").Unique(),
        field.String("vendor"),
        entjsonmap.Field("body"),
    }
}
```

## Use case 3 — Compose for finer control

```go
import "entgo.io/ent/schema/field"
import "github.com/ubgo/entkit/jsonmap"

field.JSON("metadata", jsonmap.JsonMap{}).
    Optional().
    Comment("user-controlled profile keys").
    SchemaType(map[string]string{"postgres": "jsonb"})
```

## API reference

| Function | Purpose |
|----------|---------|
| `Field(name string) ent.Field` | Returns an `ent.Field` for a `jsonmap.JsonMap` column. JSON / JSONB at the SQL layer. |

## License

Apache-2.0 — see [`../LICENSE`](../LICENSE).
