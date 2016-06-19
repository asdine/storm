package index_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/storm"
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
		assert.Equal(t, 1, countBuckets(t, idx.IndexBucket))

		err = idx.Add([]byte("hello"), []byte("id2"))
		assert.NoError(t, err)
		assert.Equal(t, 1, countBuckets(t, idx.IndexBucket))

		err = idx.Add([]byte("goodbye"), []byte("id1"))
		assert.NoError(t, err)
		assert.Equal(t, 2, countBuckets(t, idx.IndexBucket))

		err = idx.Add([]byte("hello"), []byte("id3"))
		assert.NoError(t, err)
		assert.Equal(t, 2, countBuckets(t, idx.IndexBucket))

		err = idx.RemoveID([]byte("id1"))
		assert.NoError(t, err)
		assert.Equal(t, 1, countBuckets(t, idx.IndexBucket))

		err = idx.RemoveID([]byte("id2"))
		assert.NoError(t, err)
		assert.Equal(t, 1, countBuckets(t, idx.IndexBucket))

		err = idx.RemoveID([]byte("id3"))
		assert.NoError(t, err)
		assert.Equal(t, 0, countBuckets(t, idx.IndexBucket))
		return nil
	})
}

func countBuckets(t *testing.T, bucket *bolt.Bucket) int {
	c := bucket.Cursor()
	count := 0
	for bucketName, val := c.First(); bucketName != nil; bucketName, val = c.Next() {
		if val != nil || bytes.Equal(bucketName, []byte("storm__ids")) {
			continue
		}
		count++
	}

	return count
}
