# entkit

> Custom column types and ent field helpers for the Ent ORM + gqlgen pair. One repo, many tiny independent Go modules — each pulls only the dependencies for the types you use.

`entkit` bundles three new column types plus ent field helpers for those types and for the pre-existing `ubgo/jsontype` and `ubgo/jsonslice` repos. Every type implements `database/sql/driver.Valuer`, `sql.Scanner`, and duck-typed `MarshalGQL` / `UnmarshalGQL` (so gqlgen autobinds without anyone in this repo importing gqlgen).

## Modules

### Column types (sub-modules)

| Sub-module | Type | Deps |
|------------|------|------|
| [`jsonmap`](./jsonmap) | `JsonMap` — `map[string]any` JSON object | stdlib |
| [`passwordtype`](./passwordtype) | `HashedPassword` — argon2id one-way | `github.com/ubgo/crypt` |
| [`encryptedtype`](./encryptedtype) | `EncryptedString` — AES-256-GCM, transparent CBC fallback | `github.com/ubgo/crypt` |

### Ent field helpers

| Sub-module | Wraps | Use |
|------------|-------|-----|
| [`ent_jsontype`](./ent_jsontype) | `github.com/ubgo/jsontype` | `entjsontype.Field("payload")` |
| [`ent_jsonslice`](./ent_jsonslice) | `github.com/ubgo/jsonslice` | `entjsonslice.Field("tags")` |
| [`ent_jsonmap`](./ent_jsonmap) | `entkit/jsonmap` | `entjsonmap.Field("metadata")` |
| [`ent_passwordtype`](./ent_passwordtype) | `entkit/passwordtype` | `entpasswordtype.Field("password")` |
| [`ent_encryptedtype`](./ent_encryptedtype) | `entkit/encryptedtype` | `entencryptedtype.Field("api_secret")` |

### Examples

[`examples/`](./examples) ships a runnable demo that round-trips all five types through `Value`/`Scan` and `MarshalGQL`. Run it:

```sh
cd examples && go run ./demo
```

## Why one repo with many modules

Each sub-directory is its own Go module. Consumers import only what they use — a service that only needs `JsonMap` carries zero `ubgo/crypt` and zero `entgo` in its `go.sum`. The umbrella repo simplifies issue tracking, PRs, CI, and changelog while preserving full per-feature dep isolation.

## Install

Pick the sub-modules you need:

```go
import (
    "github.com/ubgo/entkit/passwordtype"
    "github.com/ubgo/entkit/ent_passwordtype"
)
```

## License

Apache-2.0 — see [`LICENSE`](LICENSE).
