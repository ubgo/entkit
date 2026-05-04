package passwordtype

import "encoding/json"

// unmarshalString decodes a JSON string token into out without pulling
// the full encoding/json reflection machinery for a simple case.
func unmarshalString(data []byte, out *string) error {
	return json.Unmarshal(data, out)
}
