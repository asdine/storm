package storm

import (
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// One returns one record by the specified index
func (n *Node) One(fieldName string, value interface{}, to interface{}) error {
	ref := reflect.ValueOf(to)

	if !ref.IsValid() || (ref.Kind() != reflect.Ptr && structs.IsStruct(to)) {
		return ErrStructPtrNeeded
	}

	if fieldName == "" {
		return ErrNotFound
	}

	info, err := extract(to)
	if err != nil {
		return err
	}

	val, err := toBytes(value, n.s.Codec)
	if err != nil {
		return err
	}

	if n.tx != nil {
		return n.one(n.tx, fieldName, info, to, val, fieldName == info.ID.Field.Name())
	}

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.one(tx, fieldName, info, to, val, fieldName == info.ID.Field.Name())
	})
}

func (n *Node) one(tx *bolt.Tx, fieldName string, info *modelInfo, to interface{}, val []byte, skipIndex bool) error {
	bucket := n.GetBucket(tx, info.Name)
	if bucket == nil {
		return fmt.Errorf("bucket %s doesn't exist", info.Name)
	}

	var id []byte
	if !skipIndex {
		idxInfo, ok := info.Indexes[fieldName]
		if !ok {
			return ErrNotFound
		}

		idx, err := getIndex(bucket, idxInfo.Type, fieldName)
		if err != nil {
			return err
		}

		id = idx.Get(val)
	} else {
		id = val
	}

	if id == nil {
		return ErrNotFound
	}

	raw := bucket.Get(id)
	if raw == nil {
		return ErrNotFound
	}

	return n.s.Codec.Decode(raw, to)
}

// One returns one record by the specified index
func (s *DB) One(fieldName string, value interface{}, to interface{}) error {
	return s.root.One(fieldName, value, to)
}
