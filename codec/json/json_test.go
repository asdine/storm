package json

import (
	"testing"

	"github.com/AndersonBargas/rainstorm/v3/codec/internal"
)

func TestJSON(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
