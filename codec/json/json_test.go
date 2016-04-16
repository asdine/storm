package json

import (
	"testing"

	"github.com/asdine/storm/codec"
)

func TestJSON(t *testing.T) {
	codec.RountripTester(t, Codec)
}
