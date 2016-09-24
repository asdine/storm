package index_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/gob"
	"github.com/asdine/storm/index"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestUniqueIndex(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	err := db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		assert.NoError(t, err)

		idx, err := index.NewUniqueIndex(b, []byte("uindex1"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id2"))
		assert.Error(t, err)
		assert.Equal(t, index.ErrAlreadyExists, err)

		err = idx.Add(nil, []byte("id2"))
		assert.Error(t, err)
		assert.Equal(t, index.ErrNilParam, err)

		err = idx.Add([]byte("hi"), nil)
		assert.Error(t, err)
		assert.Equal(t, index.ErrNilParam, err)

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

		list, err := idx.AllRecords(nil)
		assert.NoError(t, err)
		assert.Len(t, list, 3)

		opts := index.NewOptions()
		opts.Limit = 2
		list, err = idx.AllRecords(opts)
		assert.NoError(t, err)
		assert.Len(t, list, 2)

		opts = index.NewOptions()
		opts.Skip = 2
		list, err = idx.AllRecords(opts)
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, []byte("id3"), list[0])

		opts = index.NewOptions()
		opts.Skip = 2
		opts.Limit = 1
		opts.Reverse = true
		list, err = idx.AllRecords(opts)
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, []byte("id1"), list[0])

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

func TestUniqueIndexRange(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		assert.NoError(t, err)

		idx, err := index.NewUniqueIndex(b, []byte("uindex1"))
		assert.NoError(t, err)

		for i := 0; i < 10; i++ {
			val, _ := gob.Codec.Marshal(i)
			err = idx.Add(val, val)
			assert.NoError(t, err)
		}

		min, _ := gob.Codec.Marshal(3)
		max, _ := gob.Codec.Marshal(5)
		list, err := idx.Range(min, max, nil)
		assert.Len(t, list, 3)
		assert.NoError(t, err)
		assertEncodedIntListEqual(t, []int{3, 4, 5}, list)

		min, _ = gob.Codec.Marshal(11)
		max, _ = gob.Codec.Marshal(20)
		list, err = idx.Range(min, max, nil)
		assert.Len(t, list, 0)
		assert.NoError(t, err)

		min, _ = gob.Codec.Marshal(7)
		max, _ = gob.Codec.Marshal(2)
		list, err = idx.Range(min, max, nil)
		assert.Len(t, list, 0)
		assert.NoError(t, err)

		min, _ = gob.Codec.Marshal(-5)
		max, _ = gob.Codec.Marshal(2)
		list, err = idx.Range(min, max, nil)
		assert.Len(t, list, 0)
		assert.NoError(t, err)

		min, _ = gob.Codec.Marshal(3)
		max, _ = gob.Codec.Marshal(7)
		opts := index.NewOptions()
		opts.Skip = 2
		list, err = idx.Range(min, max, opts)
		assert.Len(t, list, 3)
		assert.NoError(t, err)
		assertEncodedIntListEqual(t, []int{5, 6, 7}, list)

		opts = index.NewOptions()
		opts.Limit = 2
		list, err = idx.Range(min, max, opts)
		assert.Len(t, list, 2)
		assert.NoError(t, err)
		assertEncodedIntListEqual(t, []int{3, 4}, list)

		opts = index.NewOptions()
		opts.Reverse = true
		opts.Skip = 2
		opts.Limit = 2
		list, err = idx.Range(min, max, opts)
		assert.Len(t, list, 2)
		assert.NoError(t, err)
		assertEncodedIntListEqual(t, []int{5, 4}, list)
		return nil
	})
}

func assertEncodedIntListEqual(t *testing.T, expected []int, actual [][]byte) {
	ints := make([]int, len(actual))

	for i, e := range actual {
		err := gob.Codec.Unmarshal(e, &ints[i])
		assert.NoError(t, err)
	}

	assert.Equal(t, expected, ints)
}
