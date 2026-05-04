# ent_jsontype

> Ent field helper for [`github.com/ubgo/jsontype`](https://github.com/ubgo/jsontype). One line of schema instead of remembering the right `field.JSON(name, jsontype.JSON{})` incantation.

## Install

```sh
go get github.com/ubgo/entkit/ent_jsontype
```

## Use case 1 — Opaque JSON column

```go
import (
    "entgo.io/ent"

    entjsontype "github.com/ubgo/entkit/ent_jsontype"
)

type Event struct{ ent.Schema }

func (Event) Fields() []ent.Field {
    return []ent.Field{
        entjsontype.Field("payload"),
    }
}
```

After `task entg`, you get:

```go
client.Event.Create().
    SetPayload(jsontype.JSON(`{"k":"v"}`)).
    Save(ctx)
```

## Use case 2 — Optional + indexed (compose directly)

For column tweaks beyond the default, compose the underlying ent field call yourself:

```go
import (
    "entgo.io/ent/schema/field"

    "github.com/ubgo/jsontype"
)

func (Event) Fields() []ent.Field {
    return []ent.Field{
        field.JSON("payload", jsontype.JSON{}).
            Optional().
            Comment("opaque event payload as parsed JSON").
            SchemaType(map[string]string{"postgres": "jsonb"}),
    }
}
```

The `entjsontype.Field(name)` helper covers the common case; the full ent API is one import away.

## Use case 3 — gqlgen scalar wiring

In `gqlgen.yml`:

```yaml
models:
  JSON:
    model: github.com/ubgo/jsontype.JSON
```

In your schema:

```graphql
scalar JSON

type Event {
  id: ID!
  payload: JSON
}
```

`jsontype.JSON` already implements duck-typed `MarshalGQL`/`UnmarshalGQL`, so no resolver code is needed for the scalar.

## API reference

| Function | Purpose |
|----------|---------|
| `Field(name string) ent.Field` | Returns an `ent.Field` for a `jsontype.JSON` column. JSON / JSONB at the SQL layer. |

## License

Apache-2.0 — see [`../LICENSE`](../LICENSE).
