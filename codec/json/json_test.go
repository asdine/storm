package json

import (
	"testing"

	"github.com/AndersonBargas/rainstorm/v5/codec/internal"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}

func TestCodecName(t *testing.T) {
	require.EqualValues(t, Codec.Name(), "json")
}
