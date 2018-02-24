// Package protobuf contains a codec to encode and decode entities in Protocol Buffer
package protobuf

import (
	"errors"

	"github.com/asdine/storm/codec/json"
	"github.com/golang/protobuf/proto"
)

const name = "protobuf"

// More details on Protocol Buffers https://github.com/golang/protobuf
var (
	Codec                       = new(protobufCodec)
	errNotProtocolBufferMessage = errors.New("value isn't a Protocol Buffers Message")
)

type protobufCodec int

// Encode value with protocol buffer.
// If type isn't a Protocol buffer Message, json encoder will be used instead.
func (c protobufCodec) Marshal(v interface{}) ([]byte, error) {
	message, ok := v.(proto.Message)
	if !ok {
		// toBytes() may need to encode non-protobuf type, if that occurs use json
		return json.Codec.Marshal(v)
	}
	return proto.Marshal(message)
}

func (c protobufCodec) Unmarshal(b []byte, v interface{}) error {
	message, ok := v.(proto.Message)
	if !ok {
		// toBytes() may have encoded non-protobuf type, if that occurs use json
		return json.Codec.Unmarshal(b, v)
	}
	return proto.Unmarshal(b, message)
}

func (c protobufCodec) Name() string {
	return name
}
