package storm

import (
	"testing"

	"github.com/asdine/storm/codec/json"
	"github.com/stretchr/testify/assert"
)

type isAStringer int

func (isAStringer) String() string {
	return "I'm a stringer"
}

type isAJSONMarshaler int

func (isAJSONMarshaler) MarshalJSON() ([]byte, error) {
	return []byte("I'm a JSONMarshaler"), nil
}

func TestToBytes(t *testing.T) {
	b, err := toBytes("a string", nil, false)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a string"), b)

	b, err = toBytes(new(isAStringer), nil, false)
	assert.NoError(t, err)
	assert.Equal(t, []byte("I'm a stringer"), b)

	b, err = toBytes(new(isAJSONMarshaler), nil, false)
	assert.NoError(t, err)
	assert.Equal(t, []byte("I'm a JSONMarshaler"), b)

	b, err = toBytes(5, nil, false)
	assert.NoError(t, err)
	assert.NotNil(t, b)

	b, err = toBytes([]byte("Hey"), nil, false)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Hey"), b)
}

func TestToBytesWithCodec(t *testing.T) {
	b, err := toBytes(&SimpleUser{ID: 10, Name: "John", age: 100}, json.Codec, true)
	assert.NoError(t, err)
	assert.Equal(t, `{"ID":10,"Name":"John"}`, string(b))

	b, err = toBytes(&SimpleUser{ID: 10, Name: "John", age: 100}, json.Codec, false)
	assert.NoError(t, err)
	assert.NotEqual(t, `{"ID":10,"Name":"John"}`, string(b))

	b, err = toBytes(&SimpleUser{ID: 10, Name: "John", age: 100}, nil, false)
	assert.NoError(t, err)
	assert.NotEqual(t, `{"ID":10,"Name":"John"}`, string(b))
}
