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

func TestSave(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	err := db.Save(&SimpleUser{ID: 10, Name: "John"})
	assert.NoError(t, err)

	err = db.Save(&SimpleUser{Name: "John"})
	assert.Error(t, err)
	assert.EqualError(t, err, "id field must not be a zero value")

	err = db.Save(&UserWithNoID{Name: "John"})
	assert.Error(t, err)
	assert.EqualError(t, err, "missing struct tag id or ID field")

	err = db.Save(&UserWithIDField{ID: 10, Name: "John"})
	assert.NoError(t, err)

	u := UserWithEmbeddedIDField{}
	u.ID = 150
	u.Name = "Pete"
	u.Age = 10
	err = db.Save(&u)
	assert.NoError(t, err)

	v := UserWithIDField{ID: 10, Name: "John"}
	err = db.Save(&v)

	assert.NoError(t, err)

	db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("UserWithIDField"))
		assert.NotNil(t, bucket)

		i, err := toBytes(10)
		assert.NoError(t, err)

		val := bucket.Get(i)
		assert.NotNil(t, val)

		content, err := json.Marshal(&v)
		assert.NoError(t, err)
		assert.Equal(t, content, val)
		return nil
	})
}

type UniqueNameUser struct {
	ID   int    `storm:"id"`
	Name string `storm:"unique"`
	age  int
}

func TestSaveUnique(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	u1 := UniqueNameUser{ID: 10, Name: "John", age: 10}
	err := db.Save(&u1)
	assert.NoError(t, err)

	u2 := UniqueNameUser{ID: 11, Name: "John", age: 100}
	err = db.Save(&u2)
	assert.Error(t, err)
	assert.EqualError(t, err, "already exists")

	// same id
	u3 := UniqueNameUser{ID: 10, Name: "Jake", age: 100}
	err = db.Save(&u3)
	assert.NoError(t, err)

	db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("UniqueNameUser"))

		uniqueBucket := bucket.Bucket([]byte("Name"))
		assert.NotNil(t, uniqueBucket)

		id := uniqueBucket.Get([]byte("Jake"))
		i, err := toBytes(10)
		assert.NoError(t, err)
		assert.Equal(t, i, id)

		id = uniqueBucket.Get([]byte("John"))
		assert.Nil(t, id)
		return nil
	})
}

func TestSaveIndex(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	u1 := IndexedNameUser{ID: 10, Name: "John", age: 10}
	err := db.Save(&u1)
	assert.NoError(t, err)

	u1 = IndexedNameUser{ID: 10, Name: "John", age: 10}
	err = db.Save(&u1)
	assert.NoError(t, err)

	u2 := IndexedNameUser{ID: 11, Name: "John", age: 100}
	err = db.Save(&u2)
	assert.NoError(t, err)

	name1 := "Jake"
	name2 := "Jane"
	name3 := "James"

	for i := 0; i < 1000; i++ {
		u := IndexedNameUser{ID: i + 1}

		if i%2 == 0 {
			u.Name = name1
		} else {
			u.Name = name2
		}

		db.Save(&u)
	}

	var users []IndexedNameUser
	err = db.Find("Name", name1, &users)
	assert.NoError(t, err)
	assert.Len(t, users, 500)

	err = db.Find("Name", name2, &users)
	assert.NoError(t, err)
	assert.Len(t, users, 500)

	err = db.Find("Name", name3, &users)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	err = db.Save(nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "provided data must be a struct or a pointer to struct")
}
