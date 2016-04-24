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

		idx, err := NewUniqueIndex(b, []byte("uindex1"))
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
		assert.Equal(t, ErrNilParam, err)

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
		ids, err := idx.All([]byte("yo"), nil)
		assert.NoError(t, err)
		assert.Len(t, ids, 1)
		assert.Equal(t, []byte("id3"), ids[0])

		err = idx.RemoveID([]byte("id2"))
		assert.NoError(t, err)
		err = idx.RemoveID([]byte("id4"))
		assert.NoError(t, err)
		return nil
	})

	assert.NoError(t, err)
}

func TestListIndex(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	err := db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		assert.NoError(t, err)

		idx, err := NewListIndex(b, []byte("lindex1"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id2"))
		assert.NoError(t, err)

		err = idx.Add([]byte("goodbye"), []byte("id2"))
		assert.NoError(t, err)

		err = idx.Add(nil, []byte("id2"))
		assert.Error(t, err)
		assert.Equal(t, ErrNilParam, err)

		err = idx.Add([]byte("hi"), nil)
		assert.Error(t, err)
		assert.Equal(t, ErrNilParam, err)

		ids, err := idx.All([]byte("hello"), nil)
		assert.Len(t, ids, 1)
		assert.Equal(t, []byte("id1"), ids[0])

		ids, err = idx.All([]byte("goodbye"), nil)
		assert.Len(t, ids, 1)
		assert.Equal(t, []byte("id2"), ids[0])

		ids, err = idx.All([]byte("yo"), nil)
		assert.Nil(t, ids)

		err = idx.RemoveID([]byte("id2"))
		assert.NoError(t, err)

		ids, err = idx.All([]byte("goodbye"), nil)
		assert.Len(t, ids, 0)

		err = idx.RemoveID(nil)
		assert.NoError(t, err)

		err = idx.RemoveID([]byte("id1"))
		assert.NoError(t, err)
		err = idx.RemoveID([]byte("id2"))
		assert.NoError(t, err)
		err = idx.RemoveID([]byte("id3"))
		assert.NoError(t, err)

		ids, err = idx.All([]byte("hello"), nil)
		assert.NoError(t, err)
		assert.Nil(t, ids)

		err = idx.Add([]byte("hello"), []byte("id1"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hi"), []byte("id2"))
		assert.NoError(t, err)

		err = idx.Add([]byte("yo"), []byte("id3"))
		assert.NoError(t, err)

		err = idx.RemoveID([]byte("id2"))
		assert.NoError(t, err)

		ids, err = idx.All([]byte("hello"), nil)
		assert.Len(t, ids, 1)
		assert.Equal(t, []byte("id1"), ids[0])
		ids, err = idx.All([]byte("hi"), nil)
		assert.Len(t, ids, 0)
		ids, err = idx.All([]byte("yo"), nil)
		assert.Len(t, ids, 1)
		assert.Equal(t, []byte("id3"), ids[0])

		err = idx.RemoveID([]byte("id2"))
		assert.NoError(t, err)
		err = idx.RemoveID([]byte("id4"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hey"), []byte("id1"))
		err = idx.Add([]byte("hey"), []byte("id2"))
		err = idx.Add([]byte("hey"), []byte("id3"))
		err = idx.Add([]byte("hey"), []byte("id4"))
		ids, err = idx.All([]byte("hey"), nil)
		assert.Len(t, ids, 4)

		id := idx.Get([]byte("hey"))
		assert.Equal(t, []byte("id1"), id)

		idx.Remove([]byte("hey"))
		ids, err = idx.All([]byte("hey"), nil)
		assert.NoError(t, err)
		assert.Len(t, ids, 0)

		ids, err = idx.All([]byte("hey"), nil)
		assert.NoError(t, err)
		assert.Len(t, ids, 0)
		return nil
	})

	assert.NoError(t, err)
}
