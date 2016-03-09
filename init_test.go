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

	var u IndexedNameUser
	err := db.One("Name", "John", &u)
	assert.Error(t, err)
	assert.EqualError(t, err, "bucket IndexedNameUser doesn't exist")

	err = db.Init(&u)
	assert.NoError(t, err)

	err = db.One("Name", "John", &u)
	assert.Error(t, err)
	assert.EqualError(t, err, "not found")
}
