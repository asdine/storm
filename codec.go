package storm

import (
	"encoding/json"
)

// EncodeDecoder represents a codec used to encode and decode entities.
type EncodeDecoder interface {
	Encode(v interface{}) ([]byte, error)
	Decode(b []byte, v interface{}) error
}

// Defaults to JSON
type jsonCodec int

func (j jsonCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j jsonCodec) Decode(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}

var defaultCodec = new(jsonCodec)
