package storm

import (
	"errors"

	"github.com/boltdb/bolt"
)

func (s *Storm) addToUniqueIndex(index []byte, id []byte, key []byte, parent *bolt.Bucket) error {
	bucket, err := parent.CreateBucketIfNotExists(index)
	if err != nil {
		return err
	}

	exists := bucket.Get(key)
	if exists != nil {
		return errors.New("already exists")
	}

	return bucket.Put(key, id)
}
