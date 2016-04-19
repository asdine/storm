package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	var u IndexedNameUser
	err := db.One("Name", "John", &u)
	assert.Error(t, err)
	assert.EqualError(t, err, "bucket IndexedNameUser doesn't exist")

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
