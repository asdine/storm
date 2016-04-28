package storm

import (
	"github.com/boltdb/bolt"
)

// Set a key/value pair into a bucket
func (n *Node) Set(bucketName string, key interface{}, value interface{}) error {
	if key == nil {
		return ErrNilParam
	}

	id, err := toBytes(key, n.s.Codec)
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

	if n.tx != nil {
		return n.set(n.tx, bucketName, id, data)
	}

	return n.s.Bolt.Update(func(tx *bolt.Tx) error {
		return n.set(tx, bucketName, id, data)
	})
}

func (n *Node) set(tx *bolt.Tx, bucketName string, id, data []byte) error {
	bucket, err := n.CreateBucketIfNotExists(tx, bucketName)
	if err != nil {
		return err
	}
	return bucket.Put(id, data)
}

// Set a key/value pair into a bucket
func (s *DB) Set(bucketName string, key interface{}, value interface{}) error {
	return s.root.Set(bucketName, key, value)
}
