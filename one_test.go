package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOne(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	u := UniqueNameUser{Name: "John", ID: 10}
	err := db.Save(&u)
	assert.NoError(t, err)

	v := UniqueNameUser{}
	err = db.One("Name", "John", &v)
	assert.NoError(t, err)
	assert.Equal(t, u, v)

	for i := 0; i < 10; i++ {
		w := IndexedNameUser{Name: "John", ID: i + 1}
		err := db.Save(&w)
		assert.NoError(t, err)
	}

	x := IndexedNameUser{}
	err = db.One("Name", "John", &x)
	assert.NoError(t, err)
	assert.Equal(t, "John", x.Name)
	assert.Equal(t, 1, x.ID)
	assert.Zero(t, x.age)
	assert.True(t, x.DateOfBirth.IsZero())

	err = db.One("Name", "Mike", &x)
	assert.Error(t, err)
	assert.EqualError(t, err, "not found")

	err = db.One("", nil, &x)
	assert.Error(t, err)
	assert.EqualError(t, err, "not found")

	err = db.One("", "Mike", nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "provided target must be a pointer to struct")

	err = db.One("", nil, nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "provided target must be a pointer to struct")
}
