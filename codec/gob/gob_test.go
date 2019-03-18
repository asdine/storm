package gob

import (
	"testing"

	"github.com/asdine/storm/v2/codec/internal"
)

func TestGob(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
