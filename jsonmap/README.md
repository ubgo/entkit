# jsonmap

> A `map[string]any` you can persist to a JSON/JSONB column **and** expose as a
> gqlgen scalar — without importing gqlgen, lib/pq, or pgx.

`JsonMap` is the object-shaped peer of [`ubgo/jsonslice`](https://github.com/ubgo/jsonslice). Same three-integration pattern, swap slice for map.

| Integration | Hooks | Cost to your dependency tree |
|-------------|-------|------------------------------|
| `encoding/json` | `MarshalJSON` / `UnmarshalJSON` | stdlib |
| `database/sql` | `Value()` / `Scan()` | stdlib |
| gqlgen scalar | `MarshalGQL` / `UnmarshalGQL` | none — detected by duck typing |

## Install

```sh
go get github.com/ubgo/jsonmap
```

## Use

```go
import "github.com/ubgo/jsonmap"

type Event struct {
    ID       string
    Metadata jsonmap.JsonMap  // {"source": "stripe", "attempt": 3}
}

// Save → JSON column
db.Exec(`INSERT INTO events(id, metadata) VALUES (?, ?)`, e.ID, e.Metadata)

// Load
var loaded Event
db.QueryRow(`SELECT id, metadata FROM events WHERE id = ?`, "x").
    Scan(&loaded.ID, &loaded.Metadata)
```

In a gqlgen `gqlgen.yml`:

```yaml
models:
  JSON:
    model: github.com/ubgo/jsonmap.JsonMap
```

## Behaviour

- **`null` is preserved.** A nil map marshals to `null`; the JSON value `null`
  deserialises back to a nil map. An empty non-nil map marshals to `{}`.
- **Scan accepts `[]byte`, `string`, or `nil`.** Anything else returns an error.
  Empty bytes scan to nil.
- **`MarshalGQL` does not panic.** Encoding errors fall back to writing `{}` rather
  than panicking inside a request goroutine.

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
