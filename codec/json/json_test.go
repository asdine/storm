package json

import (
	"testing"

	"github.com/asdine/storm/codec/internal"
)

func TestJSON(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
