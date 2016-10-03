package storm

import (
	"testing"

	"github.com/asdine/storm/codec/gob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	var u IndexedNameUser
	err := db.One("Name", "John", &u)
	assert.Equal(t, ErrNotFound, err)

	err = db.Init(&u)
	assert.NoError(t, err)

	err = db.One("Name", "John", &u)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	err = db.Init(&ClassicBadTags{})
	assert.Error(t, err)
	assert.Equal(t, ErrUnknownTag, err)

	err = db.Init(10)
	assert.Error(t, err)
	assert.Equal(t, ErrBadType, err)

	err = db.Init(&ClassicNoTags{})
	assert.Error(t, err)
	assert.Equal(t, ErrNoID, err)

	err = db.Init(&struct{ ID string }{})
	assert.Error(t, err)
	assert.Equal(t, ErrNoName, err)
}

func TestInitMetadata(t *testing.T) {
	db, cleanup := createDB(t, Batch())
	defer cleanup()

	err := db.Init(new(User))
	require.NoError(t, err)
	n := db.WithCodec(gob.Codec)
	err = n.Init(new(User))
	require.Equal(t, ErrDifferentCodec, err)
}
