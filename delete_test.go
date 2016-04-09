package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	err := db.Set("files", "myfile.csv", "a,b,c,d")
	assert.NoError(t, err)
	err = db.Delete("files", "myfile.csv")
	assert.NoError(t, err)
	err = db.Delete("files", "myfile.csv")
	assert.NoError(t, err)
	err = db.Delete("i don't exist", "myfile.csv")
	assert.Equal(t, ErrNotFound, err)
	err = db.Delete("", nil)
	assert.Equal(t, ErrNotFound, err)
}
