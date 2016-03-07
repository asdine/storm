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

// NewUniqueIndex loads a UniqueIndex
func NewUniqueIndex(parent *bolt.Bucket, indexName string) (*UniqueIndex, error) {
	b, err := parent.CreateBucketIfNotExists([]byte(indexName))
	if err != nil {
		return nil, err
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
