package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestUniqueIndex(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	err := db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		assert.NoError(t, err)

		idx, err := NewUniqueIndex(b, "uindex1")
		assert.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id2"))
		assert.Error(t, err)
		assert.Equal(t, ErrAlreadyExists, err)

		err = idx.Add(nil, []byte("id2"))
		assert.Error(t, err)
		assert.Equal(t, bolt.ErrKeyRequired, err)

		err = idx.Add([]byte("hi"), nil)
		assert.Error(t, err)
		assert.Equal(t, ErrNilParam, err)

		id := idx.Get([]byte("hello"))
		assert.Equal(t, []byte("id1"), id)

		id = idx.Get([]byte("goodbye"))
		assert.Nil(t, id)

		err = idx.Remove([]byte("hello"))
		assert.NoError(t, err)

		err = idx.Remove(nil)
		assert.NoError(t, err)

		id = idx.Get([]byte("hello"))
		assert.Nil(t, id)

		err = idx.Add([]byte("hello"), []byte("id1"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hi"), []byte("id2"))
		assert.NoError(t, err)

		err = idx.Add([]byte("yo"), []byte("id3"))
		assert.NoError(t, err)

		err = idx.RemoveID([]byte("id2"))
		assert.NoError(t, err)

		id = idx.Get([]byte("hello"))
		assert.Equal(t, []byte("id1"), id)
		id = idx.Get([]byte("hi"))
		assert.Nil(t, id)
		id = idx.Get([]byte("yo"))
		assert.Equal(t, []byte("id3"), id)

		err = idx.RemoveID([]byte("id2"))
		assert.NoError(t, err)
		err = idx.RemoveID([]byte("id4"))
		assert.NoError(t, err)
		return nil
	})

	assert.NoError(t, err)
}
