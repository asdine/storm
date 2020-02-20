package msgpack

import (
	"testing"

	"github.com/AndersonBargas/rainstorm/v3/codec/internal"
)

func TestMsgpack(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
