package msgpack

import (
	"testing"

	"github.com/AndersonBargas/rainstorm/v4/codec/internal"
)

func TestMsgpack(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
