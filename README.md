# entkit

> Custom column types and ent field helpers for the **Ent ORM + gqlgen** pair. One repo, many tiny independent Go modules. Each consumer pulls only the dependencies for the types they actually import.

`entkit` ships **eight sub-modules** plus a runnable example. Three new column types, five ent field helpers (covering the three new types plus the pre-existing `ubgo/jsontype` and `ubgo/jsonslice`).

Every column type implements the same triple integration:

| Integration | Detected by |
|-------------|-------------|
| `database/sql/driver.Valuer` + `sql.Scanner` | Method signatures |
| `encoding/json.Marshaler` / `Unmarshaler` | Method signatures |
| gqlgen scalar (`MarshalGQL` / `UnmarshalGQL`) | **Duck typing** — entkit does not import gqlgen |

The third row is the headline. Consumers who don't use gqlgen never inherit a gqlgen import. Consumers who do, autobind the type as a scalar via one line in `gqlgen.yml`.

## Sub-modules

### Column types

| Sub-module | Type | What it stores | Deps |
|------------|------|---------------|------|
| [`jsonmap`](./jsonmap) | `JsonMap` | Dynamic JSON object (`{...}`) | stdlib only |
| [`passwordtype`](./passwordtype) | `HashedPassword` | One-way argon2id hash for user authentication | `github.com/ubgo/crypt` |
| [`encryptedtype`](./encryptedtype) | `EncryptedString` | Two-way AES-256-GCM (with transparent CBC fallback on reads) | `github.com/ubgo/crypt` |

Pre-existing column types in the wider `ubgo/*` family that pair with the helpers below:

| Repo | Type | What it stores |
|------|------|---------------|
| [`ubgo/jsontype`](https://github.com/ubgo/jsontype) | `JSON` | Opaque `json.RawMessage`-shaped JSON |
| [`ubgo/jsonslice`](https://github.com/ubgo/jsonslice) | `JsonSlice` | Dynamic JSON array (`[...]`) |

### Ent field helpers

| Sub-module | Wraps | One-line use |
|------------|-------|--------------|
| [`ent_jsontype`](./ent_jsontype) | `ubgo/jsontype` | `entjsontype.Field("payload")` |
| [`ent_jsonslice`](./ent_jsonslice) | `ubgo/jsonslice` | `entjsonslice.Field("tags")` |
| [`ent_jsonmap`](./ent_jsonmap) | `entkit/jsonmap` | `entjsonmap.Field("metadata")` |
| [`ent_passwordtype`](./ent_passwordtype) | `entkit/passwordtype` | `entpasswordtype.Field("password")` |
| [`ent_encryptedtype`](./ent_encryptedtype) | `entkit/encryptedtype` | `entencryptedtype.Field("api_secret")` |

### Runnable example

[`examples/`](./examples) ships a runnable demo that round-trips all five types through `Value`/`Scan` and `MarshalGQL` to prove the pattern without requiring a real database. Run:

```sh
cd examples && go run ./demo
```

## Why one repo with many sub-modules

Each sub-directory is **its own Go module**. Consumers import exactly what they use:

```go
// Service that only needs a password hash column
import (
    "github.com/ubgo/entkit/passwordtype"
    "github.com/ubgo/entkit/ent_passwordtype"
)
// → go.sum carries: ubgo/crypt, golang.org/x/crypto, entgo.io/ent
// → does NOT carry: ubgo/jsontype, ubgo/jsonslice, gqlgen
```

The umbrella simplifies issue tracking, PRs, CI, and changelog maintenance while preserving full per-feature dependency isolation. This pattern matches `entgo.io/contrib` and `kubernetes-sigs/*` at scale.

## Install matrix

Install only the sub-modules you need:

```sh
# Pick any combination
go get github.com/ubgo/entkit/jsonmap
go get github.com/ubgo/entkit/passwordtype
go get github.com/ubgo/entkit/encryptedtype

go get github.com/ubgo/entkit/ent_jsontype
go get github.com/ubgo/entkit/ent_jsonslice
go get github.com/ubgo/entkit/ent_jsonmap
go get github.com/ubgo/entkit/ent_passwordtype
go get github.com/ubgo/entkit/ent_encryptedtype
```

## Quick start — full ent + gqlgen User entity

```go
// schema/user.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"

    entencryptedtype "github.com/ubgo/entkit/ent_encryptedtype"
    entjsonmap       "github.com/ubgo/entkit/ent_jsonmap"
    entpasswordtype  "github.com/ubgo/entkit/ent_passwordtype"
)

type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("email").Unique(),

        entjsonmap.Field("metadata"),                // dynamic JSON object
        entpasswordtype.Field("password"),           // argon2id, redacts on output
        entencryptedtype.Field("api_secret").(...),  // AES-256-GCM
    }
}
```

```yaml
# gqlgen.yml
autobind:
  - github.com/ubgo/entkit/jsonmap
  - github.com/ubgo/entkit/passwordtype
  - github.com/ubgo/entkit/encryptedtype
  - github.com/ubgo/jsontype
  - github.com/ubgo/jsonslice

models:
  HashedPassword:
    model: github.com/ubgo/entkit/passwordtype.HashedPassword
  EncryptedString:
    model: github.com/ubgo/entkit/encryptedtype.EncryptedString
  JSON:
    model: github.com/ubgo/jsontype.JSON
  JSONMap:
    model: github.com/ubgo/entkit/jsonmap.JsonMap
  JSONSlice:
    model: github.com/ubgo/jsonslice.JsonSlice
```

```graphql
# schema.graphql
scalar HashedPassword
scalar EncryptedString
scalar JSON
scalar JSONMap
scalar JSONSlice

input SignupInput {
  email: String!
  password: HashedPassword!     # plaintext input, hashed before resolver sees it
  metadata: JSONMap
  apiSecret: EncryptedString    # plaintext input, encrypted at SQL layer
}

type User {
  id: ID!
  email: String!
  password: HashedPassword      # always null in output — defense in depth
  metadata: JSONMap
  apiSecret: EncryptedString    # plaintext on output (mark @internal for server-only)
}
```

## Boot wiring

`encryptedtype` requires an AES key configured once at process startup:

```go
import "github.com/ubgo/entkit/encryptedtype"

func main() {
    if err := encryptedtype.SetKey([]byte(os.Getenv("ENCRYPTION_KEY"))); err != nil {
        log.Fatal(err)
    }
    // ...
}
```

`passwordtype` uses argon2id parameters baked into [`ubgo/crypt`](https://github.com/ubgo/crypt) — no boot wiring required.

## Versioning

Each sub-module is versioned independently. Tags use the form `<sub-module>/vX.Y.Z`, e.g. `passwordtype/v0.1.0`. The umbrella repo itself does not carry a version — you install per sub-module.

For local development across multiple sub-modules, the root `go.work` stitches them together transparently.

## Repository layout

```
ubgo/entkit/
├── README.md                 ← you are here
├── LICENSE                   (Apache-2.0)
├── CHANGELOG.md              (per-sub-module entries)
├── CONTRIBUTING.md
├── NOTICE
├── go.work                   (local-dev workspace)
├── Taskfile.yml              (test/lint across all sub-modules)
│
├── jsonmap/                  ← own go.mod
├── passwordtype/             ← own go.mod
├── encryptedtype/            ← own go.mod
│
├── ent_jsontype/             ← own go.mod
├── ent_jsonslice/            ← own go.mod
├── ent_jsonmap/              ← own go.mod
├── ent_passwordtype/         ← own go.mod
├── ent_encryptedtype/        ← own go.mod
│
├── examples/                 ← runnable demo (own go.mod)
│   ├── demo/
│   └── schema/
│
└── .github/workflows/        ← matrix CI per sub-module
```

## Contributing

See [`CONTRIBUTING.md`](./CONTRIBUTING.md). Sub-modules are versioned independently; PRs should scope commits with the sub-module name (e.g. `feat(passwordtype): ...`).

## License

Apache-2.0 — see [`LICENSE`](./LICENSE).
