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
	assert.EqualError(t, err, "bucket not found")
}

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
