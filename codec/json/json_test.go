package json

import (
	"testing"

	"github.com/AndersonBargas/rainstorm/v4/codec/internal"
)

func TestJSON(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
