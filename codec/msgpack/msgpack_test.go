package msgpack

import (
	"testing"

	"github.com/asdine/storm/codec/internal"
)

func TestMsgpack(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
