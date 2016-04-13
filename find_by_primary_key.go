package storm

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Find returns one or more records by the specified index
func (s *DB) FindByPrimaryKey(value interface{}, to interface{}) error {
	d := structs.New(to)
	bucketName := d.Name()
	if bucketName == "" {
		return ErrNoName
	}
	return s.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		val, err := toBytes(value)
		if err != nil {
			return err
		}

		raw := bucket.Get(val)
		err = json.Unmarshal(raw, &to)
		if err != nil {
			return err
		}
		return nil
	})
}
