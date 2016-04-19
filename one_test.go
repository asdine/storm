package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestOne(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

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
	assert.Equal(t, ErrNotFound, err)

	err = db.One("", nil, &x)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	err = db.One("", "Mike", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrStructPtrNeeded, err)

	err = db.One("", nil, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrStructPtrNeeded, err)

	y := UniqueNameUser{Name: "Jake", ID: 200}
	err = db.Save(&y)
	assert.NoError(t, err)

	var y2 UniqueNameUser
	err = db.One("ID", 200, &y2)
	assert.NoError(t, err)
	assert.Equal(t, y, y2)

	n := NestedID{}
	n.ID = "100"
	n.Name = "John"

	err = db.Save(&n)
	assert.NoError(t, err)

	var n2 NestedID
	err = db.One("ID", "100", &n2)
	assert.NoError(t, err)
	assert.Equal(t, n, n2)
}

func TestOneNotWritable(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	err := db.Save(&User{ID: 10, Name: "John"})
	assert.NoError(t, err)

	db.Close()

	db, _ = Open(filepath.Join(dir, "storm.db"), BoltOptions(0660, &bolt.Options{
		ReadOnly: true,
	}))
	defer db.Close()

	err = db.Save(&User{ID: 20, Name: "John"})
	assert.Error(t, err)

	var u User
	err = db.One("ID", 10, &u)
	assert.NoError(t, err)
	assert.Equal(t, 10, u.ID)
	assert.Equal(t, "John", u.Name)

	err = db.One("Name", "John", &u)
	assert.NoError(t, err)
	assert.Equal(t, 10, u.ID)
	assert.Equal(t, "John", u.Name)
}
