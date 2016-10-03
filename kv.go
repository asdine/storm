package storm

import (
	"reflect"

	"github.com/boltdb/bolt"
)

// KeyValueStore can store and fetch values by key
type KeyValueStore interface {
	// Get a value from a bucket
	Get(bucketName string, key interface{}, to interface{}) error
	// Set a key/value pair into a bucket
	Set(bucketName string, key interface{}, value interface{}) error
	// Delete deletes a key from a bucket
	Delete(bucketName string, key interface{}) error
}

// Get a value from a bucket
func (n *node) Get(bucketName string, key interface{}, to interface{}) error {
	ref := reflect.ValueOf(to)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr {
		return ErrPtrNeeded
	}

	id, err := toBytes(key, n.s.codec)
	if err != nil {
		return err
	}

	return n.readTx(func(tx *bolt.Tx) error {
		return n.get(tx, bucketName, id, to)
	})
}

func (n *node) get(tx *bolt.Tx, bucketName string, id []byte, to interface{}) error {
	bucket := n.GetBucket(tx, bucketName)
	if bucket == nil {
		return ErrNotFound
	}

	raw := bucket.Get(id)
	if raw == nil {
		return ErrNotFound
	}

	return n.s.codec.Unmarshal(raw, to)
}

// Set a key/value pair into a bucket
func (n *node) Set(bucketName string, key interface{}, value interface{}) error {
	if key == nil {
		return ErrNilParam
	}

	id, err := toBytes(key, n.s.codec)
	if err != nil {
		return err
	}

	var data []byte
	if value != nil {
		data, err = n.s.codec.Marshal(value)
		if err != nil {
			return err
		}
	}

	return n.readWriteTx(func(tx *bolt.Tx) error {
		return n.set(tx, bucketName, id, data)
	})
}

func (n *node) set(tx *bolt.Tx, bucketName string, id, data []byte) error {
	bucket, err := n.CreateBucketIfNotExists(tx, bucketName)
	if err != nil {
		return err
	}

	// save node configuration in the bucket
	err = n.saveMetadata(bucket)
	if err != nil {
		return err
	}

	return bucket.Put(id, data)
}

// Delete deletes a key from a bucket
func (n *node) Delete(bucketName string, key interface{}) error {
	id, err := toBytes(key, n.s.codec)
	if err != nil {
		return err
	}

	return n.readWriteTx(func(tx *bolt.Tx) error {
		return n.delete(tx, bucketName, id)
	})
}

func (n *node) delete(tx *bolt.Tx, bucketName string, id []byte) error {
	bucket := n.GetBucket(tx, bucketName)
	if bucket == nil {
		return ErrNotFound
	}

	return bucket.Delete(id)
}

// Get a value from a bucket
func (s *DB) Get(bucketName string, key interface{}, to interface{}) error {
	return s.root.Get(bucketName, key, to)
}

// Set a key/value pair into a bucket
func (s *DB) Set(bucketName string, key interface{}, value interface{}) error {
	return s.root.Set(bucketName, key, value)
}

// Delete deletes a key from a bucket
func (s *DB) Delete(bucketName string, key interface{}) error {
	return s.root.Delete(bucketName, key)
}
