package storm

import (
	"testing"

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
	b, err := toBytes("a string")
	assert.NoError(t, err)
	assert.Equal(t, []byte("a string"), b)

	b, err = toBytes(new(isAStringer))
	assert.NoError(t, err)
	assert.Equal(t, []byte("I'm a stringer"), b)

	b, err = toBytes(new(isAJSONMarshaler))
	assert.NoError(t, err)
	assert.Equal(t, []byte("I'm a JSONMarshaler"), b)

	b, err = toBytes(5)
	assert.NoError(t, err)
	assert.NotNil(t, b)
}
