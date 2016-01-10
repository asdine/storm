package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStorm(t *testing.T) {
	db, err := New("")

	assert.Error(t, err)
	assert.Nil(t, db)

	dir, err := ioutil.TempDir(os.TempDir(), "storm")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "storm.db")
	db, err = New(file)

	assert.NoError(t, err)
	assert.Equal(t, file, db.Path)
	assert.NotNil(t, db.Bolt)
}
