package storm

import (
	"testing"

	"reflect"

	"github.com/stretchr/testify/assert"
)

func TestCodec(t *testing.T) {
	u1 := &SimpleUser{Name: "John"}
	encoded, err := defaultCodec.Encode(u1)
	assert.Nil(t, err)
	u2 := &SimpleUser{}
	err = defaultCodec.Decode(encoded, u2)
	assert.Nil(t, err)
	if !reflect.DeepEqual(u1, u2) {
		t.Fatal("Codec mismatch")
	}
}
