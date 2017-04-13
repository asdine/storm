// Package msgpack contains a codec to encode and decode entities in msgpack format
package msgpack

import (
	mp "gopkg.in/vmihailenco/msgpack.v2"
)

const name = "msgpack"

// Codec that encodes to and decodes from msgpack.
var Codec = new(msgpackCodec)

type msgpackCodec int

func (m msgpackCodec) Marshal(v interface{}) ([]byte, error) {
	return mp.Marshal(v)
}

func (m msgpackCodec) Unmarshal(b []byte, v interface{}) error {
	return mp.Unmarshal(b, v)
}

func (m msgpackCodec) Name() string {
	return name
}
