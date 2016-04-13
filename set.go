package storm

import (
	"github.com/boltdb/bolt"
)

// Set a key/value pair into a bucket
func (n *Node) Set(bucketName string, key interface{}, value interface{}) error {
	if key == nil {
		return ErrNilParam
	}

	id, err := toBytes(key)
	if err != nil {
		return err
	}

	var data []byte
	if value != nil {
		data, err = n.s.Codec.Encode(value)
		if err != nil {
			return err
		}
	}

	return n.s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := n.createBucketIfNotExists(tx, bucketName)
		if err != nil {
			return err
		}
		return bucket.Put(id, data)
	})
}

// Set a key/value pair into a bucket
func (s *DB) Set(bucketName string, key interface{}, value interface{}) error {
	return s.root.Set(bucketName, key, value)
}
