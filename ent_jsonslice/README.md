# ent_jsonslice

> Ent field helper for [`github.com/ubgo/jsonslice`](https://github.com/ubgo/jsonslice). For JSON **array** columns.

## Install

```sh
go get github.com/ubgo/entkit/ent_jsonslice
```

## Use case 1 — Tag list

```go
import (
    "entgo.io/ent"

    entjsonslice "github.com/ubgo/entkit/ent_jsonslice"
)

type Article struct{ ent.Schema }

func (Article) Fields() []ent.Field {
    return []ent.Field{
        entjsonslice.Field("tags"),
    }
}
```

```go
client.Article.Create().
    SetTags(jsonslice.JsonSlice{"go", "ent", "graphql"}).
    Save(ctx)
```

## Use case 2 — Mixed-type audit trail

The underlying type is `[]any`, so heterogeneous arrays round-trip cleanly:

```go
event.Steps = jsonslice.JsonSlice{
    "validation_passed",
    map[string]any{"step": "charge", "amount_cents": 5000},
    true,
    nil,
}
```

## Use case 3 — Compose for finer control

```go
import "entgo.io/ent/schema/field"
import "github.com/ubgo/jsonslice"

field.JSON("tags", jsonslice.JsonSlice{}).
    Optional().
    Comment("free-form tag list").
    SchemaType(map[string]string{"postgres": "jsonb"})
```

## When to use vs `entjsonmap` vs `entjsontype`

| Helper | When |
|--------|------|
| `entjsonslice.Field` | JSON array (`[...]`) — order matters, mixed types fine |
| [`entjsonmap.Field`](../ent_jsonmap) | JSON object (`{...}`) with dynamic keys |
| [`entjsontype.Field`](../ent_jsontype) | Opaque JSON — you'll route bytes elsewhere, don't care about Go shape |

## API reference

| Function | Purpose |
|----------|---------|
| `Field(name string) ent.Field` | Returns an `ent.Field` for a `jsonslice.JsonSlice` column. JSON / JSONB at the SQL layer. |

## License

Apache-2.0 — see [`../LICENSE`](../LICENSE).
