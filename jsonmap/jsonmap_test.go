package jsonmap

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJsonMap_JSONRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   JsonMap
		want string
	}{
		{"nil", nil, "null"},
		{"empty", JsonMap{}, "{}"},
		{"flat", JsonMap{"k": "v", "n": float64(1)}, ""},
		{"nested", JsonMap{"arr": []any{"x", float64(2)}, "obj": map[string]any{"k": "v"}}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if tc.want != "" && string(data) != tc.want {
				t.Fatalf("Marshal: got %s want %s", data, tc.want)
			}

			var got JsonMap
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			data2, err := json.Marshal(got)
			if err != nil {
				t.Fatalf("re-Marshal: %v", err)
			}
			if string(data2) != string(data) {
				t.Fatalf("round trip diverged:\n  first:  %s\n  second: %s", data, data2)
			}
		})
	}
}

func TestJsonMap_UnmarshalJSON_Invalid(t *testing.T) {
	var j JsonMap
	if err := j.UnmarshalJSON([]byte(`["a"]`)); err == nil {
		t.Fatal("expected error for array input")
	}
	if err := j.UnmarshalJSON([]byte(`not json`)); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestJsonMap_SQLValue(t *testing.T) {
	var nilMap JsonMap
	v, err := nilMap.Value()
	if err != nil {
		t.Fatalf("Value(nil): %v", err)
	}
	if v != nil {
		t.Fatalf("nil map Value() = %v, want nil", v)
	}

	m := JsonMap{"k": "v"}
	v, err = m.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	got, ok := v.([]byte)
	if !ok {
		t.Fatalf("Value type = %T, want []byte", v)
	}
	if string(got) != `{"k":"v"}` {
		t.Fatalf("Value = %s", got)
	}
}

func TestJsonMap_SQLScan(t *testing.T) {
	cases := []struct {
		name    string
		src     any
		wantNil bool
		want    string
		wantErr bool
	}{
		{"nil src", nil, true, "", false},
		{"bytes", []byte(`{"k":"v"}`), false, `{"k":"v"}`, false},
		{"string", `{"k":"v"}`, false, `{"k":"v"}`, false},
		{"empty bytes", []byte{}, true, "", false},
		{"int unsupported", 42, false, "", true},
		{"bad json", []byte(`{`), false, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var j JsonMap
			err := j.Scan(tc.src)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Scan: %v", err)
			}
			if tc.wantNil {
				if j != nil {
					t.Fatalf("want nil, got %v", j)
				}
				return
			}
			data, _ := json.Marshal(j)
			if string(data) != tc.want {
				t.Fatalf("Scan got %s want %s", data, tc.want)
			}
		})
	}
}

func TestJsonMap_MarshalGQL(t *testing.T) {
	var buf bytes.Buffer
	JsonMap{"k": "v"}.MarshalGQL(&buf)
	if buf.String() != `{"k":"v"}` {
		t.Fatalf("MarshalGQL = %s", buf.String())
	}
}

func TestJsonMap_UnmarshalGQL(t *testing.T) {
	cases := []struct {
		name    string
		in      any
		want    string
		wantErr bool
	}{
		{"nil", nil, "null", false},
		{"map", map[string]any{"k": "v"}, `{"k":"v"}`, false},
		{"raw json object", json.RawMessage(`{"k":"v"}`), `{"k":"v"}`, false},
		{"raw json array", json.RawMessage(`["a"]`), "", true},
		{"unsupported", make(chan int), "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var j JsonMap
			err := j.UnmarshalGQL(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalGQL: %v", err)
			}
			data, _ := json.Marshal(j)
			if string(data) != tc.want {
				t.Fatalf("got %s want %s", data, tc.want)
			}
		})
	}
}
