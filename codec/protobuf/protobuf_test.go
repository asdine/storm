package protobuf

import (
	"testing"

	"github.com/asdine/storm/codec/internal"
	"github.com/stretchr/testify/assert"
)

func TestProtobuf(t *testing.T) {
	u1 := SimpleUser{
		Id: uint64(1),
		Name: "John",
	}
	u2 := SimpleUser{}
	internal.RoundtripTester(t, Codec, &u1, &u2)
	assert.True(t, u1.Id == u2.Id)
}
