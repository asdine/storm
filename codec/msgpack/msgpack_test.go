package msgpack

import (
	"testing"

	"github.com/asdine/storm/v3/codec/internal"
)

func TestMsgpack(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
