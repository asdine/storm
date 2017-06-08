package ulid

import (
	"testing"

	"github.com/oklog/ulid"

	"encoding"

	"github.com/stretchr/testify/require"
)

func TestUlid(t *testing.T) {
	assert := require.New(t)
	var current, last ulid.ULID
	next := New(nil)
	for i := 0; i < 10; i++ {
		v, err := next(nil)
		assert.NoError(err)
		assert.NotNil(v)
		bm := v.(encoding.BinaryMarshaler)
		bmb, err := bm.MarshalBinary()
		assert.NoError(err)
		assert.Len(bmb, 16)
		current = v.(ulid.ULID)
		assert.Len(current, 16)
		assert.NotEqual(current, last)
		last = current
	}
}
