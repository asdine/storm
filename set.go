package storm

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

// Set a key/value pair into a bucket
func (s *DB) Set(bucketName string, key interface{}, value interface{}) error {
	if key == nil {
		return ErrNilParam
	}

	id, err := toBytes(key)
	if err != nil {
		return err
	}

	var data []byte
	if value != nil {
		data, err = json.Marshal(value)
		if err != nil {
			return err
		}
	}

	return s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return bucket.Put(id, data)
	})
}
