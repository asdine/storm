package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStorm(t *testing.T) {
	db, err := Open("")

	assert.Error(t, err)
	assert.Nil(t, db)

	dir, err := ioutil.TempDir(os.TempDir(), "storm")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "storm.db")
	db, err = Open(file)

	assert.NoError(t, err)
	assert.Equal(t, file, db.Path)
	assert.NotNil(t, db.Bolt)
}

func TestNewStormWithOptions(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := OpenWithOptions(filepath.Join(dir, "storm.db"), 0600, nil)
	defer db.Close()

	err := db.Save(&SimpleUser{ID: 10})
	assert.NoError(t, err)
}
