package storm

import (
	"reflect"

	"github.com/boltdb/bolt"
)

// Get a value from a bucket
func (n *Node) Get(bucketName string, key interface{}, to interface{}) error {
	ref := reflect.ValueOf(to)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr {
		return ErrPtrNeeded
	}

	id, err := toBytes(key)
	if err != nil {
		return err
	}

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		bucket := n.getBucket(tx, bucketName)
		if bucket == nil {
			return ErrNotFound
		}

		raw := bucket.Get(id)
		if raw == nil {
			return ErrNotFound
		}

		return n.s.Codec.Decode(raw, to)
	})
}

// Get a value from a bucket
func (s *DB) Get(bucketName string, key interface{}, to interface{}) error {
	return s.root.Get(bucketName, key, to)
}
