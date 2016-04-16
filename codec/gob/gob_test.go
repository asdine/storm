package json

import (
	"testing"

	"github.com/asdine/storm/codec"
)

func TestGob(t *testing.T) {
	codec.RountripTester(t, Codec)
}
