package json

import (
	"testing"

	"github.com/asdine/storm/v2/codec/internal"
)

func TestJSON(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
