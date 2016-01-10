package storm

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

type User struct {
	ID   int    `storm:"id"`
	Name string `storm:"unique"`
	age  int
}

type BadFriend struct {
	Name string
}

type GoodFriend struct {
	ID   int
	Name string `storm:"id"`
}

type GoodOtherFriend struct {
	ID   int
	Name string
}

func TestSave(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := New(filepath.Join(dir, "storm.db"))

	u1 := User{ID: 10, Name: "John"}
	err := db.Save(&u1)

	u2 := User{Name: "John"}
	err = db.Save(&u2)
	assert.Error(t, err)
	assert.EqualError(t, err, "id field must not be a zero value")

	u3 := BadFriend{Name: "John"}
	err = db.Save(&u3)
	assert.Error(t, err)
	assert.EqualError(t, err, "missing struct tag id")

	u4 := GoodFriend{ID: 10, Name: "John"}
	err = db.Save(&u4)
	assert.NoError(t, err)

	u5 := GoodOtherFriend{ID: 10, Name: "John"}
	err = db.Save(&u5)
	assert.NoError(t, err)

	db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("user"))
		assert.NotNil(t, bucket)

		i, err := toBytes(10)
		assert.NoError(t, err)

		val := bucket.Get(i)
		assert.NotNil(t, val)

		content, err := json.Marshal(&u5)
		assert.NoError(t, err)
		assert.Equal(t, content, val)
		return nil
	})
}

func TestSaveUnique(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := New(filepath.Join(dir, "storm.db"))

	u1 := User{ID: 10, Name: "John", age: 10}
	err := db.Save(&u1)
	assert.NoError(t, err)

	u2 := User{ID: 11, Name: "John", age: 100}
	err = db.Save(&u2)
	assert.Error(t, err)
	assert.EqualError(t, err, "already exists")

	db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("user"))

		uniqueBucket := bucket.Bucket([]byte("Name"))
		assert.NotNil(t, uniqueBucket)

		id := uniqueBucket.Get([]byte("John"))
		i, err := toBytes(10)
		assert.NoError(t, err)
		assert.Equal(t, i, id)

		return nil
	})
}
