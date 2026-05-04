# jsonmap-ent

> Ent field helper for [`github.com/ubgo/entkit/jsonmap`](https://github.com/ubgo/entkit/jsonmap).

## Install

```sh
go get github.com/ubgo/entkit/jsonmap-ent
```

## Use

```go
import (
    "entgo.io/ent"

    entjsonmap "github.com/ubgo/entkit/jsonmap-ent"
)

type Event struct{ ent.Schema }

func (Event) Fields() []ent.Field {
    return []ent.Field{
        entjsonmap.Field("metadata"),
    }
}
```

For finer control compose `field.JSON("metadata", jsonmap.JsonMap{}).Optional()...` directly.

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
