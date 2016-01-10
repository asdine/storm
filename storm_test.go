package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	ID   int `storm:"id"`
	Name string
	age  int
}

type BadFriend struct {
	ID   int
	Name string
}

type GoodFriend struct {
	ID   int
	Name string `storm:"id"`
}

func TestStorm(t *testing.T) {
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

	u1 := User{ID: 10, Name: "John"}
	err = db.Save(&u1)

	u2 := User{Name: "John"}
	err = db.Save(&u2)
	assert.Error(t, err)
	assert.EqualError(t, err, "id field must not be a zero value")

	u3 := BadFriend{ID: 10, Name: "John"}
	err = db.Save(&u3)
	assert.Error(t, err)
	assert.EqualError(t, err, "missing struct tag id")

	u4 := GoodFriend{ID: 10, Name: "John"}
	err = db.Save(&u4)
	assert.NoError(t, err)
}
