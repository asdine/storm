package q

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type StringAndBytes struct {
	A string
	B []byte
	C int
}

func TestRe(t *testing.T) {
	a := StringAndBytes{
		A: "ABC",
		B: []byte("234"),
	}

	b := StringAndBytes{
		A: "123",
		B: []byte("DEF"),
	}

	q := Re("A", "\\d+")
	am, err := q.Match(&a)
	require.NoError(t, err)
	require.False(t, am)

	bm, err := q.Match(&b)
	require.NoError(t, err)
	require.True(t, bm)

	q = Re("B", "\\d+")
	am, err = q.Match(&a)
	require.NoError(t, err)
	require.True(t, am)

	bm, err = q.Match(&b)
	require.NoError(t, err)
	require.False(t, bm)

	// Field C is int, regexp not supported
	q = Re("C", "\\d+")
	_, err = q.Match(&b)
	require.Error(t, err)

	// Invalid regexp
	q = Re("A", "\\d++")
	_, err = q.Match(&b)
	require.Error(t, err)

}
