package gob

import (
	"testing"

	"github.com/AndersonBargas/rainstorm/v5/codec/internal"
	"github.com/stretchr/testify/require"
)

func TestGob(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}

func TestCodecName(t *testing.T) {
	require.EqualValues(t, Codec.Name(), "gob")
}
