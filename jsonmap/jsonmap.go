// Package jsonmap is a JSON-backed map type that round-trips through SQL
// columns and GraphQL scalars without dragging gqlgen or any database
// driver into the dependency graph.
//
// The core type is [JsonMap], an alias for map[string]any that:
//
//   - Marshals/unmarshals as a JSON object via the standard encoding/json
//     interfaces.
//   - Implements sql.Scanner and driver.Valuer so it can be stored in a
//     JSON / JSONB column directly.
//   - Implements MarshalGQL/UnmarshalGQL so gqlgen will pick it up as a
//     custom scalar via duck typing — no gqlgen import is required, and
//     consumers that don't use gqlgen pay nothing.
//
// The zero value is a nil map. nil is treated as the JSON value null
// when scanning; on marshal a nil [JsonMap] is rendered as null too. An
// empty non-nil map marshals to {}.
package jsonmap

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// JsonMap is a JSON object keyed by string with arbitrary JSON values.
type JsonMap map[string]any

// MarshalJSON implements [json.Marshaler]. A nil map marshals to the JSON
// value null; an empty non-nil map marshals to {}.
func (j JsonMap) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]any(j))
}

// UnmarshalJSON implements [json.Unmarshaler]. The JSON value null
// deserialises to a nil map.
func (j *JsonMap) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*j = nil
		return nil
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("jsonmap: unmarshal: %w", err)
	}
	*j = raw
	return nil
}

// Value implements [driver.Valuer]. The zero / nil map persists as SQL
// NULL; otherwise the map is encoded as a JSON object.
func (j JsonMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(map[string]any(j))
}

// Scan implements [sql.Scanner]. Accepts []byte, string, or nil. NULL
// scans into a nil map.
func (j *JsonMap) Scan(src any) error {
	if src == nil {
		*j = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("jsonmap: cannot scan %T", src)
	}
	if len(data) == 0 {
		*j = nil
		return nil
	}
	return j.UnmarshalJSON(data)
}

// MarshalGQL is the gqlgen-compatible marshal hook. gqlgen detects this
// method by name; we therefore need not import gqlgen. Errors are rare
// (the map already round-tripped through JSON to reach a resolver) and
// would indicate a programmer error encoding an unsupported type into the
// map — those are surfaced by writing an empty object, never panicking
// inside a request goroutine.
func (j JsonMap) MarshalGQL(w io.Writer) {
	data, err := json.Marshal(map[string]any(j))
	if err != nil {
		_, _ = w.Write([]byte("{}"))
		return
	}
	_, _ = w.Write(data)
}

// UnmarshalGQL is the gqlgen-compatible unmarshal hook. It accepts the
// already-parsed JSON value gqlgen passes (a map[string]any for a JSON
// object, or a JSON-encoded byte slice / string), and decodes into the
// map.
func (j *JsonMap) UnmarshalGQL(v any) error {
	if v == nil {
		*j = nil
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		*j = m
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("jsonmap: %T is not a JSON object", v)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return errors.New("jsonmap: value is not a JSON object")
	}
	*j = raw
	return nil
}
