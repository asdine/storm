package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemove(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	u1 := IndexedNameUser{ID: 10, Name: "John", age: 10}
	err := db.Save(&u1)
	assert.NoError(t, err)

	err = db.Remove(&u1)
	assert.NoError(t, err)

	err = db.Remove(&u1)
	assert.EqualError(t, err, "not found")

	u2 := IndexedNameUser{}
	err = db.Get("IndexedNameUser", 10, &u2)
	assert.EqualError(t, err, "not found")
}
