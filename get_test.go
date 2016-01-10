package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOneByIndex(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := New(filepath.Join(dir, "storm.db"))

	u := UniqueNameUser{Name: "John", ID: 10}
	err := db.Save(&u)
	assert.NoError(t, err)

	v := UniqueNameUser{}
	err = db.OneByIndex("Name", "John", &v)
	assert.NoError(t, err)
	assert.Equal(t, u, v)

	for i := 0; i < 10; i++ {
		w := IndexedNameUser{Name: "John", ID: i + 1}
		err := db.Save(&w)
		assert.NoError(t, err)
	}

	x := IndexedNameUser{}
	err = db.OneByIndex("Name", "John", &x)
	assert.NoError(t, err)
	assert.Equal(t, IndexedNameUser{Name: "John", ID: 1}, x)

	err = db.OneByIndex("Name", "Mike", &x)
	assert.Error(t, err)
	assert.EqualError(t, err, "not found")
}
