package storm

import (
	"encoding/json"
	"io/ioutil"
	"net/mail"
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

type SimpleUser struct {
	ID   int `storm:"id"`
	Name string
	age  int
}

type UserWithNoID struct {
	Name string
}

type UserWithIDField struct {
	ID   int
	Name string
}

type UserWithEmbeddedIDField struct {
	UserWithIDField `storm:"inline"`
	Age             int
}

func TestSave(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	err := db.Save(&SimpleUser{ID: 10, Name: "John"})

	err = db.Save(&SimpleUser{Name: "John"})
	assert.Error(t, err)
	assert.EqualError(t, err, "id field must not be a zero value")

	err = db.Save(&UserWithNoID{Name: "John"})
	assert.Error(t, err)
	assert.EqualError(t, err, "missing struct tag id")

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

type IndexedNameUser struct {
	ID   int    `storm:"id"`
	Name string `storm:"index"`
	age  int
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

	db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("IndexedNameUser"))

		listBucket := bucket.Bucket([]byte("Name"))
		assert.NotNil(t, listBucket)

		raw := listBucket.Get([]byte("John"))
		assert.NotNil(t, raw)

		var list [][]byte

		err = json.Unmarshal(raw, &list)
		assert.NoError(t, err)
		assert.Len(t, list, 2)

		id1, err := toBytes(u1.ID)
		assert.NoError(t, err)
		id2, err := toBytes(u2.ID)
		assert.NoError(t, err)

		assert.Equal(t, id1, list[0])
		assert.Equal(t, id2, list[1])

		return nil
	})
}

func TestSet(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	err := db.Set("b1", 10, 10)
	assert.NoError(t, err)
	err = db.Set("b1", "best friend's mail", &mail.Address{Name: "Gandalf", Address: "gandalf@lorien.ma"})
	assert.NoError(t, err)
	err = db.Set("b2", []byte("i'm already a slice of bytes"), "a value")
	assert.NoError(t, err)
	err = db.Set("b2", []byte("i'm already a slice of bytes"), nil)
	assert.NoError(t, err)
	err = db.Set("b1", 0, 100)
	assert.NoError(t, err)
	err = db.Set("b1", nil, 100)
	assert.Error(t, err)

	db.Bolt.View(func(tx *bolt.Tx) error {
		b1 := tx.Bucket([]byte("b1"))
		assert.NotNil(t, b1)
		b2 := tx.Bucket([]byte("b2"))
		assert.NotNil(t, b2)

		k1, err := toBytes(10)
		assert.NoError(t, err)
		val := b1.Get(k1)
		assert.NotNil(t, val)

		k2 := []byte("best friend's mail")
		val = b1.Get(k2)
		assert.NotNil(t, val)

		k3, err := toBytes(0)
		assert.NoError(t, err)
		val = b1.Get(k3)
		assert.NotNil(t, val)

		return nil
	})
}
