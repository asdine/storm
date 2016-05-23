package protobuf

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/asdine/storm/codec/internal"
	"github.com/stretchr/testify/assert"
)

func TestProtobuf(t *testing.T) {
	u1 := SimpleUser{
		Id: proto.Uint64(1),
		Name: proto.String("John"),
	}
	u2 := SimpleUser{}
	internal.RoundtripTester(t, Codec, &u1, &u2)
	assert.True(t, u1.GetId() == u2.GetId())
}
