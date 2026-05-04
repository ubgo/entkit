# jsonslice-ent

> Ent field helper for [`github.com/ubgo/jsonslice`](https://github.com/ubgo/jsonslice).

## Install

```sh
go get github.com/ubgo/jsonslice-ent
```

## Use

```go
import (
    "entgo.io/ent"

    entjsonslice "github.com/ubgo/jsonslice-ent"
)

type Event struct{ ent.Schema }

func (Event) Fields() []ent.Field {
    return []ent.Field{
        entjsonslice.Field("tags"),
    }
}
```

For finer control compose `field.JSON("tags", jsonslice.JsonSlice{}).Optional()...` directly.

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
