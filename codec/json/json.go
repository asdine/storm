package json

import (
	"encoding/json"
)

// Codec that encodes to and decodes from JSON.
var Codec = new(jsonCodec)

type jsonCodec int

func (j jsonCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j jsonCodec) Decode(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}
