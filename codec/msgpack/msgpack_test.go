package msgpack

import (
	"testing"

	"github.com/AndersonBargas/rainstorm/v5/codec/internal"
	"github.com/stretchr/testify/require"
)

func TestMsgpack(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}

func TestCodecName(t *testing.T) {
	require.EqualValues(t, Codec.Name(), "msgpack")
}
