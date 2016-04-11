package storm

import (
	"bytes"

	"github.com/boltdb/bolt"
)

const indexPrefix = "__storm_index_"

// Index interface
type Index interface {
	Add(value []byte, targetID []byte) error
	Remove(value []byte) error
	RemoveID(id []byte) error
	Get(value []byte) []byte
	All(value []byte) ([][]byte, error)
	AllRecords() ([][]byte, error)
	Range([]byte, []byte) ([][]byte, error)
}

// NewUniqueIndex loads a UniqueIndex
func NewUniqueIndex(parent *bolt.Bucket, indexName []byte) (*UniqueIndex, error) {
	var err error
	b := parent.Bucket(indexName)
	if b == nil {
		if !parent.Writable() {
			return nil, ErrIndexNotFound
		}
		b, err = parent.CreateBucket(indexName)
		if err != nil {
			return nil, err
		}
	}

	return &UniqueIndex{
		IndexBucket: b,
		Parent:      parent,
	}, nil
}

// UniqueIndex is an index that references unique values and the corresponding ID.
type UniqueIndex struct {
	Parent      *bolt.Bucket
	IndexBucket *bolt.Bucket
}

// Add a value to the unique index
func (idx *UniqueIndex) Add(value []byte, targetID []byte) error {
	if value == nil || len(value) == 0 {
		return ErrNilParam
	}
	if targetID == nil || len(targetID) == 0 {
		return ErrNilParam
	}

	exists := idx.IndexBucket.Get(value)
	if exists != nil {
		if bytes.Equal(exists, targetID) {
			return nil
		}
		return ErrAlreadyExists
	}

	return idx.IndexBucket.Put(value, targetID)
}

// Remove a value from the unique index
func (idx *UniqueIndex) Remove(value []byte) error {
	return idx.IndexBucket.Delete(value)
}

// RemoveID removes an ID from the unique index
func (idx *UniqueIndex) RemoveID(id []byte) error {
	c := idx.IndexBucket.Cursor()

	for val, ident := c.First(); val != nil; val, ident = c.Next() {
		if bytes.Equal(ident, id) {
			return idx.Remove(val)
		}
	}
	return nil
}

// Get the id corresponding to the given value
func (idx *UniqueIndex) Get(value []byte) []byte {
	return idx.IndexBucket.Get(value)
}

// Get the ids corresponding to the given range of values
func (idx *UniqueIndex) Range(min []byte, max []byte) ([][]byte, error) {
	var list [][]byte
	c := idx.IndexBucket.Cursor()
	for val, id := c.Seek(min); val != nil && bytes.Compare(val, max) <= 0; val, id = c.Next() {
		list = append(list, id)
	}
	return list, nil

}

// All returns all the ids corresponding to the given value
func (idx *UniqueIndex) All(value []byte) ([][]byte, error) {
	id := idx.IndexBucket.Get(value)
	if id != nil {
		return [][]byte{id}, nil
	}

	return nil, nil
}

// AllRecords returns all the IDs of this index
func (idx *UniqueIndex) AllRecords() ([][]byte, error) {
	c := idx.IndexBucket.Cursor()

	var list [][]byte

	for val, ident := c.First(); val != nil; val, ident = c.Next() {
		list = append(list, ident)
	}
	return list, nil
}

// first returns the first ID of this index
func (idx *UniqueIndex) first() []byte {
	c := idx.IndexBucket.Cursor()

	for val, ident := c.First(); val != nil; val, ident = c.Next() {
		return ident
	}
	return nil
}

// NewListIndex loads a ListIndex
func NewListIndex(parent *bolt.Bucket, indexName []byte) (*ListIndex, error) {
	var err error
	b := parent.Bucket(indexName)
	if b == nil {
		if !parent.Writable() {
			return nil, ErrIndexNotFound
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
		if val != nil {
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

// Get the ids corresponding to the given range of values
func (idx *ListIndex) Range(min []byte, max []byte) ([][]byte, error) {
	var list [][]byte
	c := idx.IndexBucket.Cursor()
	for val, _ := c.Seek(min); val != nil && bytes.Compare(val, max) <= 0; val, _ = c.Next() {
		uni := idx.IndexBucket.Bucket(val)
		uni.ForEach(func(k, v []byte) error {
			list = append(list, k)
			return nil
		})
	}
	return list, nil
}

// All the IDs corresponding to the given value
func (idx *ListIndex) All(value []byte) ([][]byte, error) {
	uni, err := NewUniqueIndex(idx.IndexBucket, value)
	if err != nil {
		return nil, err
	}
	return uni.AllRecords()
}

// AllRecords returns all the IDs of this index
func (idx *ListIndex) AllRecords() ([][]byte, error) {
	c := idx.IndexBucket.Cursor()

	var list [][]byte

	for bucketName, val := c.First(); bucketName != nil; bucketName, val = c.Next() {
		if val != nil || bytes.Equal(bucketName, []byte("storm__ids")) {
			continue
		}

		uni, err := NewUniqueIndex(idx.IndexBucket, bucketName)
		if err != nil {
			return nil, err
		}

		all, err := uni.AllRecords()
		if err != nil {
			return nil, err
		}
		list = append(list, all...)
	}

	return list, nil
}

func getIndex(bucket *bolt.Bucket, idxKind string, fieldName string) (Index, error) {
	var idx Index
	var err error

	switch idxKind {
	case tagUniqueIdx:
		idx, err = NewUniqueIndex(bucket, []byte(indexPrefix+fieldName))
	case tagIdx:
		idx, err = NewListIndex(bucket, []byte(indexPrefix+fieldName))
	default:
		err = ErrBadIndexType
	}

	return idx, err
}
