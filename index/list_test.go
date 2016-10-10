package index_test

import (
	"bytes"
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

func TestListIndex(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	err := db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		assert.NoError(t, err)

		idx, err := index.NewListIndex(b, []byte("lindex1"))
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
		assert.Equal(t, index.ErrNilParam, err)

		err = idx.Add([]byte("hi"), nil)
		assert.Error(t, err)
		assert.Equal(t, index.ErrNilParam, err)

		ids, err := idx.All([]byte("hello"), nil)
		assert.NoError(t, err)
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

		opts := index.NewOptions()
		opts.Limit = 1
		ids, err = idx.All([]byte("hey"), opts)
		assert.Len(t, ids, 1)

		opts = index.NewOptions()
		opts.Skip = 2
		ids, err = idx.All([]byte("hey"), opts)
		assert.Len(t, ids, 2)

		opts = index.NewOptions()
		opts.Skip = 2
		opts.Limit = 3
		opts.Reverse = true
		ids, err = idx.All([]byte("hey"), opts)
		assert.Len(t, ids, 2)
		assert.Equal(t, []byte("id2"), ids[0])

		id := idx.Get([]byte("hey"))
		assert.Equal(t, []byte("id1"), id)

		err = idx.Remove([]byte("hey"))
		assert.NoError(t, err)
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

func TestListIndexAddRemoveID(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		assert.NoError(t, err)

		idx, err := index.NewListIndex(b, []byte("lindex1"))
		assert.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		assert.NoError(t, err)
		assert.Equal(t, 1, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("hello"), []byte("id2"))
		assert.NoError(t, err)
		assert.Equal(t, 2, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("goodbye"), []byte("id1"))
		assert.NoError(t, err)
		assert.Equal(t, 2, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("hello"), []byte("id3"))
		assert.NoError(t, err)
		assert.Equal(t, 3, countItems(t, idx.IndexBucket))

		err = idx.RemoveID([]byte("id1"))
		assert.NoError(t, err)
		assert.Equal(t, 2, countItems(t, idx.IndexBucket))

		err = idx.RemoveID([]byte("id2"))
		assert.NoError(t, err)
		assert.Equal(t, 1, countItems(t, idx.IndexBucket))

		err = idx.RemoveID([]byte("id3"))
		assert.NoError(t, err)
		assert.Equal(t, 0, countItems(t, idx.IndexBucket))
		return nil
	})
}

func TestListIndexAllRecords(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		assert.NoError(t, err)

		idx, err := index.NewListIndex(b, []byte("lindex1"))
		assert.NoError(t, err)

		ids, err := idx.AllRecords(nil)
		assert.NoError(t, err)
		assert.Len(t, ids, 0)

		err = idx.Add([]byte("goodbye"), []byte("id2"))
		assert.NoError(t, err)
		assert.Equal(t, 1, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("goodbye"), []byte("id1"))
		assert.NoError(t, err)
		assert.Equal(t, 2, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("hello"), []byte("id4"))
		assert.NoError(t, err)
		assert.Equal(t, 3, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("hello"), []byte("id3"))
		assert.NoError(t, err)
		assert.Equal(t, 4, countItems(t, idx.IndexBucket))

		ids, err = idx.AllRecords(nil)
		assert.NoError(t, err)
		assert.Len(t, ids, 4)
		assert.Equal(t, []byte("id1"), ids[0])
		assert.Equal(t, []byte("id2"), ids[1])
		assert.Equal(t, []byte("id3"), ids[2])
		assert.Equal(t, []byte("id4"), ids[3])

		err = idx.RemoveID([]byte("id1"))
		assert.NoError(t, err)
		assert.Equal(t, 3, countItems(t, idx.IndexBucket))

		ids, err = idx.AllRecords(nil)
		assert.NoError(t, err)
		assert.Len(t, ids, 3)
		assert.Equal(t, []byte("id2"), ids[0])

		err = idx.Add([]byte("goodbye"), []byte("id1"))
		assert.NoError(t, err)
		assert.Equal(t, 4, countItems(t, idx.IndexBucket))

		opts := index.NewOptions()
		opts.Limit = 1
		ids, err = idx.AllRecords(opts)
		assert.Len(t, ids, 1)

		opts = index.NewOptions()
		opts.Skip = 2
		ids, err = idx.AllRecords(opts)
		assert.Len(t, ids, 2)

		opts = index.NewOptions()
		opts.Skip = 2
		opts.Limit = 3
		opts.Reverse = true
		ids, err = idx.AllRecords(opts)
		assert.NoError(t, err)
		assert.Len(t, ids, 2)
		assert.Equal(t, []byte("id2"), ids[0])

		return nil
	})
}

func TestListIndexRange(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		assert.NoError(t, err)

		idx, err := index.NewListIndex(b, []byte("index1"))
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

func countItems(t *testing.T, bucket *bolt.Bucket) int {
	c := bucket.Cursor()
	count := 0
	for k, id := c.First(); k != nil; k, id = c.Next() {
		if id == nil || bytes.Equal(k, []byte("storm__ids")) {
			continue
		}
		count++
	}

	return count
}
