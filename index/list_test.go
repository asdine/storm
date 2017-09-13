package index_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/gob"
	"github.com/asdine/storm/index"
	"github.com/coreos/bbolt"
	"github.com/stretchr/testify/require"
)

func TestListIndex(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	err := db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		require.NoError(t, err)

		idx, err := index.NewListIndex(b, []byte("lindex1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id2"))
		require.NoError(t, err)

		err = idx.Add([]byte("goodbye"), []byte("id2"))
		require.NoError(t, err)

		err = idx.Add(nil, []byte("id2"))
		require.Error(t, err)
		require.Equal(t, index.ErrNilParam, err)

		err = idx.Add([]byte("hi"), nil)
		require.Error(t, err)
		require.Equal(t, index.ErrNilParam, err)

		ids, err := idx.All([]byte("hello"), nil)
		require.NoError(t, err)
		require.Len(t, ids, 1)
		require.Equal(t, []byte("id1"), ids[0])

		ids, err = idx.All([]byte("goodbye"), nil)
		require.Len(t, ids, 1)
		require.Equal(t, []byte("id2"), ids[0])

		ids, err = idx.All([]byte("yo"), nil)
		require.Nil(t, ids)

		err = idx.RemoveID([]byte("id2"))
		require.NoError(t, err)

		ids, err = idx.All([]byte("goodbye"), nil)
		require.Len(t, ids, 0)

		err = idx.RemoveID(nil)
		require.NoError(t, err)

		err = idx.RemoveID([]byte("id1"))
		require.NoError(t, err)
		err = idx.RemoveID([]byte("id2"))
		require.NoError(t, err)
		err = idx.RemoveID([]byte("id3"))
		require.NoError(t, err)

		ids, err = idx.All([]byte("hello"), nil)
		require.NoError(t, err)
		require.Nil(t, ids)

		err = idx.Add([]byte("hello"), []byte("id1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hi"), []byte("id2"))
		require.NoError(t, err)

		err = idx.Add([]byte("yo"), []byte("id3"))
		require.NoError(t, err)

		err = idx.RemoveID([]byte("id2"))
		require.NoError(t, err)

		ids, err = idx.All([]byte("hello"), nil)
		require.Len(t, ids, 1)
		require.Equal(t, []byte("id1"), ids[0])
		ids, err = idx.All([]byte("hi"), nil)
		require.Len(t, ids, 0)
		ids, err = idx.All([]byte("yo"), nil)
		require.Len(t, ids, 1)
		require.Equal(t, []byte("id3"), ids[0])

		err = idx.RemoveID([]byte("id2"))
		require.NoError(t, err)
		err = idx.RemoveID([]byte("id4"))
		require.NoError(t, err)

		err = idx.Add([]byte("hey"), []byte("id1"))
		err = idx.Add([]byte("hey"), []byte("id2"))
		err = idx.Add([]byte("hey"), []byte("id3"))
		err = idx.Add([]byte("hey"), []byte("id4"))
		ids, err = idx.All([]byte("hey"), nil)
		require.Len(t, ids, 4)

		opts := index.NewOptions()
		opts.Limit = 1
		ids, err = idx.All([]byte("hey"), opts)
		require.Len(t, ids, 1)

		opts = index.NewOptions()
		opts.Skip = 2
		ids, err = idx.All([]byte("hey"), opts)
		require.Len(t, ids, 2)

		opts = index.NewOptions()
		opts.Skip = 2
		opts.Limit = 3
		opts.Reverse = true
		ids, err = idx.All([]byte("hey"), opts)
		require.Len(t, ids, 2)
		require.Equal(t, []byte("id2"), ids[0])

		id := idx.Get([]byte("hey"))
		require.Equal(t, []byte("id1"), id)

		err = idx.Remove([]byte("hey"))
		require.NoError(t, err)
		ids, err = idx.All([]byte("hey"), nil)
		require.NoError(t, err)
		require.Len(t, ids, 0)

		ids, err = idx.All([]byte("hey"), nil)
		require.NoError(t, err)
		require.Len(t, ids, 0)
		return nil
	})

	require.NoError(t, err)
}

func TestListIndexReverse(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	err := db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		require.NoError(t, err)

		idx, err := index.NewListIndex(b, []byte("lindex1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		require.NoError(t, err)

		opts := index.NewOptions()
		ids, err := idx.All([]byte("hello"), opts)
		require.Len(t, ids, 1)
		require.Equal(t, []byte("id1"), ids[0])

		opts = index.NewOptions()
		opts.Reverse = true
		ids, err = idx.All([]byte("hello"), opts)
		require.Len(t, ids, 1)
		require.Equal(t, []byte("id1"), ids[0])

		err = idx.Add([]byte("hello"), []byte("id2"))
		require.NoError(t, err)

		opts = index.NewOptions()
		opts.Reverse = true
		ids, err = idx.All([]byte("hello"), opts)
		require.Len(t, ids, 2)
		require.Equal(t, []byte("id2"), ids[0])
		require.Equal(t, []byte("id1"), ids[1])
		return nil
	})

	require.NoError(t, err)
}

func TestListIndexAddRemoveID(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		require.NoError(t, err)

		idx, err := index.NewListIndex(b, []byte("lindex1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		require.NoError(t, err)
		require.Equal(t, 1, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("hello"), []byte("id2"))
		require.NoError(t, err)
		require.Equal(t, 2, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("goodbye"), []byte("id1"))
		require.NoError(t, err)
		require.Equal(t, 2, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("hello"), []byte("id3"))
		require.NoError(t, err)
		require.Equal(t, 3, countItems(t, idx.IndexBucket))

		err = idx.RemoveID([]byte("id1"))
		require.NoError(t, err)
		require.Equal(t, 2, countItems(t, idx.IndexBucket))

		err = idx.RemoveID([]byte("id2"))
		require.NoError(t, err)
		require.Equal(t, 1, countItems(t, idx.IndexBucket))

		err = idx.RemoveID([]byte("id3"))
		require.NoError(t, err)
		require.Equal(t, 0, countItems(t, idx.IndexBucket))
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
		require.NoError(t, err)

		idx, err := index.NewListIndex(b, []byte("lindex1"))
		require.NoError(t, err)

		ids, err := idx.AllRecords(nil)
		require.NoError(t, err)
		require.Len(t, ids, 0)

		err = idx.Add([]byte("goodbye"), []byte("id2"))
		require.NoError(t, err)
		require.Equal(t, 1, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("goodbye"), []byte("id1"))
		require.NoError(t, err)
		require.Equal(t, 2, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("hello"), []byte("id4"))
		require.NoError(t, err)
		require.Equal(t, 3, countItems(t, idx.IndexBucket))

		err = idx.Add([]byte("hello"), []byte("id3"))
		require.NoError(t, err)
		require.Equal(t, 4, countItems(t, idx.IndexBucket))

		ids, err = idx.AllRecords(nil)
		require.NoError(t, err)
		require.Len(t, ids, 4)
		require.Equal(t, []byte("id1"), ids[0])
		require.Equal(t, []byte("id2"), ids[1])
		require.Equal(t, []byte("id3"), ids[2])
		require.Equal(t, []byte("id4"), ids[3])

		err = idx.RemoveID([]byte("id1"))
		require.NoError(t, err)
		require.Equal(t, 3, countItems(t, idx.IndexBucket))

		ids, err = idx.AllRecords(nil)
		require.NoError(t, err)
		require.Len(t, ids, 3)
		require.Equal(t, []byte("id2"), ids[0])

		err = idx.Add([]byte("goodbye"), []byte("id1"))
		require.NoError(t, err)
		require.Equal(t, 4, countItems(t, idx.IndexBucket))

		opts := index.NewOptions()
		opts.Limit = 1
		ids, err = idx.AllRecords(opts)
		require.Len(t, ids, 1)

		opts = index.NewOptions()
		opts.Skip = 2
		ids, err = idx.AllRecords(opts)
		require.Len(t, ids, 2)

		opts = index.NewOptions()
		opts.Skip = 2
		opts.Limit = 3
		opts.Reverse = true
		ids, err = idx.AllRecords(opts)
		require.NoError(t, err)
		require.Len(t, ids, 2)
		require.Equal(t, []byte("id2"), ids[0])

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
		require.NoError(t, err)

		idx, err := index.NewListIndex(b, []byte("index1"))
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			val, _ := gob.Codec.Marshal(i)
			err = idx.Add(val, val)
			require.NoError(t, err)
		}

		min, _ := gob.Codec.Marshal(3)
		max, _ := gob.Codec.Marshal(5)
		list, err := idx.Range(min, max, nil)
		require.Len(t, list, 3)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{3, 4, 5}, list)

		min, _ = gob.Codec.Marshal(11)
		max, _ = gob.Codec.Marshal(20)
		list, err = idx.Range(min, max, nil)
		require.Len(t, list, 0)
		require.NoError(t, err)

		min, _ = gob.Codec.Marshal(7)
		max, _ = gob.Codec.Marshal(2)
		list, err = idx.Range(min, max, nil)
		require.Len(t, list, 0)
		require.NoError(t, err)

		min, _ = gob.Codec.Marshal(-5)
		max, _ = gob.Codec.Marshal(2)
		list, err = idx.Range(min, max, nil)
		require.Len(t, list, 0)
		require.NoError(t, err)

		min, _ = gob.Codec.Marshal(3)
		max, _ = gob.Codec.Marshal(7)
		opts := index.NewOptions()
		opts.Skip = 2
		list, err = idx.Range(min, max, opts)
		require.Len(t, list, 3)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{5, 6, 7}, list)

		opts = index.NewOptions()
		opts.Limit = 2
		list, err = idx.Range(min, max, opts)
		require.Len(t, list, 2)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{3, 4}, list)

		opts = index.NewOptions()
		opts.Reverse = true
		opts.Skip = 2
		opts.Limit = 2
		list, err = idx.Range(min, max, opts)
		require.Len(t, list, 2)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{5, 4}, list)
		return nil
	})
}

func TestListIndexPrefix(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		require.NoError(t, err)

		idx, err := index.NewListIndex(b, []byte("lindex1"))
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			val := []byte(fmt.Sprintf("a%d", i%2))
			id := []byte(fmt.Sprintf("%d", i))
			err = idx.Add(val, id)
			require.NoError(t, err)
		}

		for i := 10; i < 20; i++ {
			val := []byte(fmt.Sprintf("b%d", i%2))
			id := []byte(fmt.Sprintf("%d", i))
			err = idx.Add(val, id)
			require.NoError(t, err)
		}

		list, err := idx.Prefix([]byte("a"), nil)
		require.Len(t, list, 10)
		require.NoError(t, err)
		require.Equal(t, []byte("0"), list[0])
		require.Equal(t, []byte("9"), list[9])

		list, err = idx.Prefix([]byte("b"), nil)
		require.Len(t, list, 10)
		require.NoError(t, err)
		require.Equal(t, []byte("10"), list[0])
		require.Equal(t, []byte("19"), list[9])

		opts := index.NewOptions()
		opts.Reverse = true
		list, err = idx.Prefix([]byte("a"), opts)
		require.Len(t, list, 10)
		require.NoError(t, err)
		require.Equal(t, []byte("9"), list[0])
		require.Equal(t, []byte("0"), list[9])

		opts = index.NewOptions()
		opts.Reverse = true
		list, err = idx.Prefix([]byte("b"), opts)
		require.Len(t, list, 10)
		require.NoError(t, err)
		require.Equal(t, []byte("19"), list[0])
		require.Equal(t, []byte("10"), list[9])

		opts = index.NewOptions()
		opts.Skip = 9
		opts.Limit = 5
		list, err = idx.Prefix([]byte("a"), opts)
		require.Len(t, list, 1)
		require.NoError(t, err)
		require.Equal(t, []byte("9"), list[0])

		opts = index.NewOptions()
		opts.Reverse = true
		opts.Skip = 9
		opts.Limit = 5
		list, err = idx.Prefix([]byte("a"), opts)
		require.Len(t, list, 1)
		require.NoError(t, err)
		require.Equal(t, []byte("0"), list[0])
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
