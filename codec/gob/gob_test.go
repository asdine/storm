package gob

import (
	"testing"

	"github.com/asdine/storm/codec/internal"
)

func TestGob(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
