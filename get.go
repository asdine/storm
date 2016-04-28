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

	id, err := toBytes(key, n.s.Codec)
	if err != nil {
		return err
	}

	if n.tx != nil {
		return n.get(n.tx, bucketName, id, to)
	}

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.get(tx, bucketName, id, to)
	})
}

func (n *Node) get(tx *bolt.Tx, bucketName string, id []byte, to interface{}) error {
	bucket := n.GetBucket(tx, bucketName)
	if bucket == nil {
		return ErrNotFound
	}

	raw := bucket.Get(id)
	if raw == nil {
		return ErrNotFound
	}

	return n.s.Codec.Decode(raw, to)
}

// Get a value from a bucket
func (s *DB) Get(bucketName string, key interface{}, to interface{}) error {
	return s.root.Get(bucketName, key, to)
}
