package storm

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"sort"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Index interface
type Index interface {
	Add(value []byte, targetID []byte) error
	Remove(value []byte) error
	RemoveID(id []byte) error
	Get(value []byte) []byte
	All(value []byte) ([][]byte, error)
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

// All returns all the ids corresponding to the given value
func (idx *UniqueIndex) All(value []byte) ([][]byte, error) {
	id := idx.IndexBucket.Get(value)
	if id != nil {
		return [][]byte{id}, nil
	}

	return nil, nil
}

// allRecords returns all the IDs of this index
func (idx *UniqueIndex) allRecords() [][]byte {
	c := idx.IndexBucket.Cursor()

	var list [][]byte

	for val, ident := c.First(); val != nil; val, ident = c.Next() {
		list = append(list, ident)
	}
	return list
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

	return nil
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
func (idx *ListIndex) All(value []byte) ([][]byte, error) {
	uni, err := NewUniqueIndex(idx.IndexBucket, value)
	if err != nil {
		return nil, err
	}
	return uni.allRecords(), nil
}

func (s *DB) addToUniqueIndex(index []byte, id []byte, key []byte, parent *bolt.Bucket) error {
	bucket, err := parent.CreateBucketIfNotExists(index)
	if err != nil {
		return err
	}

	exists := bucket.Get(key)
	if exists != nil {
		if bytes.Equal(exists, id) {
			return nil
		}
		return errors.New("already exists")
	}

	return bucket.Put(key, id)
}

func (s *DB) addToListIndex(idx []byte, id []byte, key []byte, parent *bolt.Bucket) error {
	bucket, err := parent.CreateBucketIfNotExists(idx)
	if err != nil {
		return err
	}

	var list [][]byte

	raw := bucket.Get(key)
	if raw != nil {
		err = json.Unmarshal(raw, &list)
		if err != nil {
			return err
		}
	}

	list = append(list, id)
	sort.Sort(index(list))
	raw, err = json.Marshal(list)
	if err != nil {
		return err
	}

	return bucket.Put(key, raw)
}

func (s *DB) deleteOldIndexes(parent *bolt.Bucket, id []byte, indexes []*structs.Field, unique bool) error {
	raw := parent.Get(id)
	if raw == nil {
		return nil
	}

	var content map[string]interface{}
	err := json.Unmarshal(raw, &content)
	if err != nil {
		return err
	}

	for _, field := range indexes {
		bucket := parent.Bucket([]byte(field.Name()))
		if bucket == nil {
			continue
		}

		f, ok := content[field.Name()]
		if !ok || !reflect.ValueOf(f).IsValid() {
			continue
		}

		key, err := toBytes(f)
		if err != nil {
			return err
		}

		if !unique {
			raw = bucket.Get(key)
			if raw == nil {
				continue
			}

			var list [][]byte
			err := json.Unmarshal(raw, &list)
			if err != nil {
				return err
			}

			for i, ident := range list {
				if bytes.Equal(ident, id) {
					list = append(list[0:i], list[i+1:]...)
					break
				}
			}

			raw, err = json.Marshal(list)
			if err != nil {
				return err
			}

			err = bucket.Put(key, raw)
		} else {
			err = bucket.Delete(key)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

type index [][]byte

func (s index) Len() int           { return len(s) }
func (s index) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s index) Less(i, j int) bool { return bytes.Compare(s[i], s[j]) == -1 }
