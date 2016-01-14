package storm

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
)

func (s *DB) addToUniqueIndex(index []byte, id []byte, key []byte, parent *bolt.Bucket) error {
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

func (s *DB) addToListIndex(index []byte, id []byte, key []byte, parent *bolt.Bucket) error {
	bucket, err := parent.CreateBucketIfNotExists(index)
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
	raw, err = json.Marshal(&list)
	if err != nil {
		return err
	}

	return bucket.Put(key, raw)
}
