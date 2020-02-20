package gob

import (
	"testing"

	"github.com/AndersonBargas/rainstorm/v3/codec/internal"
)

func TestGob(t *testing.T) {
	internal.RoundtripTester(t, Codec)
}
