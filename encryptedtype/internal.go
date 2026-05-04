package encryptedtype

import "encoding/json"

func unmarshalString(data []byte, out *string) error {
	return json.Unmarshal(data, out)
}

func jsonString(s string) ([]byte, error) {
	return json.Marshal(s)
}
