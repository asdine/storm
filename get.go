package storm

import (
	"encoding/json"
	"reflect"

	"github.com/boltdb/bolt"
)

// Get a value from a bucket
func (s *DB) Get(bucketName string, key interface{}, to interface{}) error {
	ref := reflect.ValueOf(to)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr {
		return ErrPtrNeeded
	}

	id, err := toBytes(key)
	if err != nil {
		return err
	}

	return s.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return ErrNotFound
		}

		raw := bucket.Get(id)
		if raw == nil {
			return ErrNotFound
		}

		return json.Unmarshal(raw, to)
	})
}
