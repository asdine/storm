package storm

import (
	"testing"

	"github.com/asdine/storm/codec/gob"
	"github.com/asdine/storm/codec/json"
	"github.com/stretchr/testify/assert"
)

func TestToBytes(t *testing.T) {
	b, err := toBytes([]byte("a slice of bytes"), gob.Codec)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a slice of bytes"), b)

	b, err = toBytes("a string", gob.Codec)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a string"), b)

	b, err = toBytes(5, gob.Codec)
	assert.NoError(t, err)
	assert.NotNil(t, b)

	b, err = toBytes([]byte("Hey"), gob.Codec)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Hey"), b)
}

func TestToBytesWithCodec(t *testing.T) {
	b, err := toBytes([]byte("a slice of bytes"), json.Codec)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a slice of bytes"), b)

	b, err = toBytes("a string", json.Codec)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a string"), b)

	b, err = toBytes(&SimpleUser{ID: 10, Name: "John", age: 100}, json.Codec)
	assert.NoError(t, err)
	assert.Equal(t, `{"ID":10,"Name":"John"}`, string(b))
}
