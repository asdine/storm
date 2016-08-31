// Package protobuf contains a codec to encode and decode entities in Protocol Buffer
package protobuf

import (
	"errors"

	"github.com/asdine/storm/codec/json"
	"github.com/golang/protobuf/proto"
)

// More details on Protocol Buffers https://github.com/golang/protobuf
var (
	Codec                       = new(protobufCodec)
	errNotProtocolBufferMessage = errors.New("value isn't a Protocol Buffers Message")
)

type protobufCodec int

// Encode value with protocol buffer.
// If type isn't a Protocol buffer Message, gob encoder will be used instead.
func (c protobufCodec) Encode(v interface{}) ([]byte, error) {
	message, ok := v.(proto.Message)
	if !ok {
		// toBytes() may need to encode non-protobuf type, if that occurs use json
		return json.Codec.Encode(v)
	}
	return proto.Marshal(message)
}

func (c protobufCodec) Decode(b []byte, v interface{}) error {
	message, ok := v.(proto.Message)
	if !ok {
		return errNotProtocolBufferMessage
	}
	return proto.Unmarshal(b, message)
}
