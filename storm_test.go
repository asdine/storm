package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorm(t *testing.T) {
	db, err := New("")

	assert.Error(t, err)
	assert.Nil(t, db)

	dir, err := ioutil.TempDir(os.TempDir(), "storm")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err = New(filepath.Join(dir, "storm.db"))

	assert.NoError(t, err)
	assert.Equal(t, "storm.db", db.Path)
	assert.NotNil(t, db.Session)
}
