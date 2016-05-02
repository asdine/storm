package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/codec/gob"
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
	assert.Equal(t, ErrZeroID, err)

	err = db.Save(&ClassicBadTags{ID: "id", PublicField: 100})
	assert.Error(t, err)
	assert.Equal(t, ErrUnknownTag, err)

	err = db.Save(&UserWithNoID{Name: "John"})
	assert.Error(t, err)
	assert.Equal(t, ErrNoID, err)

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

	w := UserWithEmbeddedField{}
	w.ID = 150
	w.Name = "John"
	err = db.Save(&w)
	assert.NoError(t, err)

	db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("UserWithIDField"))
		assert.NotNil(t, bucket)

		i, err := toBytes(10, gob.Codec)
		assert.NoError(t, err)

		val := bucket.Get(i)
		assert.NotNil(t, val)

		content, err := db.Codec.Encode(&v)
		assert.NoError(t, err)
		assert.Equal(t, content, val)
		return nil
	})
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
	assert.Equal(t, ErrAlreadyExists, err)

	// same id
	u3 := UniqueNameUser{ID: 10, Name: "Jake", age: 100}
	err = db.Save(&u3)
	assert.NoError(t, err)

	db.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("UniqueNameUser"))

		uniqueBucket := bucket.Bucket([]byte(indexPrefix + "Name"))
		assert.NotNil(t, uniqueBucket)

		id := uniqueBucket.Get([]byte("Jake"))
		i, err := toBytes(10, gob.Codec)
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
	assert.Equal(t, ErrBadType, err)
}

func TestSaveEmptyValues(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	u := User{
		ID: 10,
	}
	err := db.Save(&u)
	assert.NoError(t, err)

	var v User
	err = db.One("ID", 10, &v)
	assert.NoError(t, err)
	assert.Equal(t, 10, v.ID)

	u.Name = "John"
	u.Slug = "john"
	err = db.Save(&u)
	assert.NoError(t, err)

	err = db.One("Name", "John", &v)
	assert.NoError(t, err)
	assert.Equal(t, "John", v.Name)
	assert.Equal(t, "john", v.Slug)
	err = db.One("Slug", "john", &v)
	assert.NoError(t, err)
	assert.Equal(t, "John", v.Name)
	assert.Equal(t, "john", v.Slug)

	u.Name = ""
	u.Slug = ""
	err = db.Save(&u)
	assert.NoError(t, err)

	err = db.One("Name", "John", &v)
	assert.Error(t, err)
	err = db.One("Slug", "john", &v)
	assert.Error(t, err)
}

func TestSaveAutoIncrement(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"), AutoIncrement())
	defer db.Close()

	for i := 1; i < 10; i++ {
		s := SimpleUser{Name: "John"}
		err := db.Save(&s)
		assert.NoError(t, err)
		assert.Equal(t, i, s.ID)
	}

	u := UserWithUint64IDField{Name: "John"}
	err := db.Save(&u)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), u.ID)
	v := UserWithUint64IDField{}
	err = db.One("ID", uint64(1), &v)
	assert.NoError(t, err)
	assert.Equal(t, u, v)

	us := UserWithStringIDField{Name: "John"}
	err = db.Save(&us)
	assert.Error(t, err)
	assert.Equal(t, ErrZeroID, err)
}

func TestSaveDifferentBucketRoot(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"), AutoIncrement())
	defer db.Close()

	assert.Len(t, db.rootBucket, 0)

	dbSub := db.From("sub")

	assert.NotEqual(t, dbSub, db)
	assert.Len(t, dbSub.rootBucket, 1)

	err := db.Save(&User{ID: 10, Name: "John"})
	assert.NoError(t, err)
	err = dbSub.Save(&User{ID: 11, Name: "Paul"})
	assert.NoError(t, err)

	var (
		john User
		paul User
	)

	err = db.One("Name", "John", &john)
	assert.NoError(t, err)
	err = db.One("Name", "Paul", &paul)
	assert.Error(t, err)

	err = dbSub.One("Name", "Paul", &paul)
	assert.NoError(t, err)
	err = dbSub.One("Name", "John", &john)
	assert.Error(t, err)
}
