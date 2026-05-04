# jsonmap

> A `map[string]any` column type for SQL (JSON / JSONB) and gqlgen scalars. Stdlib only.

`JsonMap` is the object-shaped peer of [`ubgo/jsonslice`](https://github.com/ubgo/jsonslice) (which handles JSON arrays). Three integrations packed into one tiny type:

| Integration | Hooks | Cost to your dependency tree |
|-------------|-------|------------------------------|
| `encoding/json` | `MarshalJSON` / `UnmarshalJSON` | stdlib |
| `database/sql` | `Value()` / `Scan()` | stdlib |
| gqlgen scalar | `MarshalGQL` / `UnmarshalGQL` | none — detected by duck typing |

The third row is the headline. gqlgen recognises any type with the right `MarshalGQL(io.Writer)` and `UnmarshalGQL(any) error` method signatures, so `JsonMap` exposes those methods *without* importing `github.com/99designs/gqlgen`. Consumers who don't use gqlgen pay nothing.

## When to use

| Use this | When you need |
|----------|---------------|
| `jsonmap.JsonMap` | A JSON **object** column (`{...}`) — shape is dynamic per row |
| [`jsonslice.JsonSlice`](https://github.com/ubgo/jsonslice) | A JSON **array** column (`[...]`) |
| [`jsontype.JSON`](https://github.com/ubgo/jsontype) | An opaque JSON column where you don't care about Go-side shape; you'll route the bytes elsewhere |

If you need a typed object with known fields, define a Go struct and use `json.RawMessage` or `field.JSON("metadata", &MyStruct{})` directly — `JsonMap` is for cases where the keys are user-controlled or vary by row.

## Install

```sh
go get github.com/ubgo/entkit/jsonmap
```

## Use case 1 — User profile metadata

A user profile carries arbitrary additional attributes that vary per integration (Stripe customer IDs, Slack user IDs, feature flag overrides, etc.). Storing each as its own column would require a migration every time a new integration ships.

```go
import "github.com/ubgo/entkit/jsonmap"

type User struct {
    ID       string
    Email    string
    Metadata jsonmap.JsonMap   // {"stripe_id": "cus_...", "slack_id": "U..."}
}

// Save
u.Metadata["stripe_id"] = "cus_abc123"
db.Exec(`UPDATE users SET metadata = ? WHERE id = ?`, u.Metadata, u.ID)

// Load
var u User
db.QueryRow(`SELECT id, email, metadata FROM users WHERE id = ?`, id).
    Scan(&u.ID, &u.Email, &u.Metadata)

stripeID, _ := u.Metadata["stripe_id"].(string)
```

## Use case 2 — Webhook event payload

Inbound webhook events from third parties carry vendor-specific payloads. You want to persist the parsed body for replay and auditing without losing fidelity.

```go
type WebhookEvent struct {
    ID       string
    Vendor   string
    Received time.Time
    Body     jsonmap.JsonMap
}

// Receive
var body jsonmap.JsonMap
if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
    http.Error(w, "bad json", 400)
    return
}
db.Exec(`INSERT INTO webhook_events(id, vendor, received, body) VALUES (?, ?, ?, ?)`,
    eventID, "stripe", time.Now(), body)
```

## Use case 3 — Feature flag overrides

Ship per-user feature flag overrides that engineers can flip in admin tooling without a migration.

```go
type FeatureOverride struct {
    UserID string
    Flags  jsonmap.JsonMap   // {"new_dashboard": true, "rollout_pct": 50}
}

if u.Flags["new_dashboard"] == true {
    return RenderNewDashboard(...)
}
```

## Use case 4 — gqlgen scalar (server)

Use it as a GraphQL scalar without importing gqlgen in your column-type module.

```yaml
# gqlgen.yml
models:
  JSON:
    model: github.com/ubgo/entkit/jsonmap.JsonMap
```

```graphql
scalar JSON

input UpdateProfile {
  metadata: JSON
}

type User {
  id: ID!
  metadata: JSON
}
```

```go
// resolver — the gqlgen runtime calls JsonMap.UnmarshalGQL for you
func (r *mutationResolver) UpdateProfile(ctx context.Context, input UpdateProfile) (*User, error) {
    return userSvc.SetMetadata(ctx, input.Metadata)
}
```

## Use case 5 — Storing in raw `database/sql`

```go
m := jsonmap.JsonMap{"theme": "dark", "tier": 2}

// Postgres jsonb column
_, err := db.Exec(`INSERT INTO settings(user_id, prefs) VALUES ($1, $2)`, userID, m)

// Reading back
var loaded jsonmap.JsonMap
err = db.QueryRow(`SELECT prefs FROM settings WHERE user_id = $1`, userID).Scan(&loaded)
```

## API reference

| Function / method | Purpose |
|---|---|
| `JsonMap` (type) | `map[string]any` alias, ready to use |
| `(j JsonMap) MarshalJSON() ([]byte, error)` | nil → `null`, otherwise JSON object |
| `(j *JsonMap) UnmarshalJSON([]byte) error` | `null` → nil; object → map |
| `(j JsonMap) Value() (driver.Value, error)` | nil → SQL NULL; otherwise JSON-encoded bytes |
| `(j *JsonMap) Scan(any) error` | accepts `[]byte`, `string`, or `nil` |
| `(j JsonMap) MarshalGQL(io.Writer)` | duck-typed gqlgen marshal |
| `(j *JsonMap) UnmarshalGQL(any) error` | accepts `map[string]any`, `[]byte`, `string`, raw JSON |

## Behaviour notes

- **`null` is preserved.** A nil map marshals to `null`; the JSON value `null` deserialises back to a nil map. An empty non-nil map marshals to `{}`.
- **Scan accepts `[]byte`, `string`, or `nil`.** Anything else returns an error. Empty bytes scan to nil.
- **`MarshalGQL` does not panic.** Encoding errors fall back to writing `{}` rather than panicking inside a request goroutine.
- **Scanning a JSON array fails.** `JsonMap` is for objects; use `JsonSlice` for arrays.

## License

Apache-2.0 — see [`../LICENSE`](../LICENSE).
