package sereal

import (
	"github.com/Sereal/Sereal/Go/sereal"
)

// The Sereal codec has some interesting features, one of them being
// serialization of object references, including circular references.
// See https://github.com/Sereal/Sereal
var Codec = new(serealCodec)

type serealCodec int

func (c serealCodec) Encode(v interface{}) ([]byte, error) {
	return sereal.Marshal(v)
}

func (c serealCodec) Decode(b []byte, v interface{}) error {
	return sereal.Unmarshal(b, v)
}
