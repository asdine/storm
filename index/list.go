package index

import (
	"bytes"

	"github.com/boltdb/bolt"
)

// NewListIndex loads a ListIndex
func NewListIndex(parent *bolt.Bucket, indexName []byte) (*ListIndex, error) {
	var err error
	b := parent.Bucket(indexName)
	if b == nil {
		if !parent.Writable() {
			return nil, ErrNotFound
		}
		b, err = parent.CreateBucket(indexName)
		if err != nil {
			return nil, err
		}
	}

	ids, err := NewUniqueIndex(b, []byte("storm__ids"))
	if err != nil {
		return nil, err
	}

	return &ListIndex{
		IndexBucket: b,
		Parent:      parent,
		IDs:         ids,
	}, nil
}

// ListIndex is an index that references values and the corresponding IDs.
type ListIndex struct {
	Parent      *bolt.Bucket
	IndexBucket *bolt.Bucket
	IDs         *UniqueIndex
}

// Add a value to the list index
func (idx *ListIndex) Add(value []byte, targetID []byte) error {
	if value == nil || len(value) == 0 {
		return ErrNilParam
	}
	if targetID == nil || len(targetID) == 0 {
		return ErrNilParam
	}

	oldValue := idx.IDs.Get(targetID)
	if oldValue != nil {
		uni, err := NewUniqueIndex(idx.IndexBucket, oldValue)
		if err != nil {
			return err
		}

		err = uni.Remove(targetID)
		if err != nil {
			return err
		}

		err = idx.IDs.Remove(targetID)
		if err != nil {
			return err
		}
	}

	uni, err := NewUniqueIndex(idx.IndexBucket, value)
	if err != nil {
		return err
	}

	err = uni.Add(targetID, targetID)
	if err != nil {
		return err
	}

	return idx.IDs.Add(targetID, value)
}

// Remove a value from the unique index
func (idx *ListIndex) Remove(value []byte) error {
	err := idx.IDs.RemoveID(value)
	if err != nil {
		return err
	}
	return idx.IndexBucket.DeleteBucket(value)
}

// RemoveID removes an ID from the list index
func (idx *ListIndex) RemoveID(targetID []byte) error {
	c := idx.IndexBucket.Cursor()

	for bucketName, val := c.First(); bucketName != nil; bucketName, val = c.Next() {
		if val != nil || bytes.Equal(bucketName, []byte("storm__ids")) {
			continue
		}

		uni, err := NewUniqueIndex(idx.IndexBucket, bucketName)
		if err != nil {
			return err
		}

		err = uni.Remove(targetID)
		if err != nil {
			return err
		}

		cd := uni.IndexBucket.Cursor()
		empty := true
		for k, _ := cd.First(); k != nil; k, _ = cd.Next() {
			empty = false
			break
		}

		if empty {
			err = idx.IndexBucket.DeleteBucket(bucketName)
			if err != nil {
				return err
			}
		}
	}

	return idx.IDs.Remove(targetID)
}

// Get the first ID corresponding to the given value
func (idx *ListIndex) Get(value []byte) []byte {
	uni, err := NewUniqueIndex(idx.IndexBucket, value)
	if err != nil {
		return nil
	}
	return uni.first()
}

// All the IDs corresponding to the given value
func (idx *ListIndex) All(value []byte, opts *Options) ([][]byte, error) {
	uni, err := NewUniqueIndex(idx.IndexBucket, value)
	if err != nil {
		return nil, err
	}
	return uni.AllRecords(opts)
}

// AllRecords returns all the IDs of this index
func (idx *ListIndex) AllRecords(opts *Options) ([][]byte, error) {
	var list [][]byte

	c := idx.IndexBucket.Cursor()

	for bucketName, val := c.First(); bucketName != nil; bucketName, val = c.Next() {
		if val != nil || bytes.Equal(bucketName, []byte("storm__ids")) {
			continue
		}

		uni, err := NewUniqueIndex(idx.IndexBucket, bucketName)
		if err != nil {
			return nil, err
		}

		all, err := uni.AllRecords(opts)
		if err != nil {
			return nil, err
		}
		list = append(list, all...)
	}

	return list, nil
}

// Range returns the ids corresponding to the given range of values
func (idx *ListIndex) Range(min []byte, max []byte, opts *Options) ([][]byte, error) {
	var list [][]byte

	c := idx.IndexBucket.Cursor()

	for bucketName, val := c.Seek(min); bucketName != nil && bytes.Compare(bucketName, max) <= 0; bucketName, val = c.Next() {
		if val != nil || bytes.Equal(bucketName, []byte("storm__ids")) {
			continue
		}

		uni, err := NewUniqueIndex(idx.IndexBucket, bucketName)
		if err != nil {
			return nil, err
		}

		all, err := uni.AllRecords(opts)
		if err != nil {
			return nil, err
		}

		list = append(list, all...)
	}

	return list, nil
}
