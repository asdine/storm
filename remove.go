package storm

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Remove removes a structure from the associated bucket
func (s *DB) Remove(data interface{}) error {
	if !structs.IsStruct(data) {
		return ErrBadType
	}

	info, err := extract(data)
	if err != nil {
		return err
	}

	id, err := toBytes(info.ID.Value())
	if err != nil {
		return err
	}

	return s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(info.Name))
		if bucket == nil {
			return fmt.Errorf("bucket %s doesn't exist", info.Name)
		}

		for fieldName, idxInfo := range info.Indexes {
			idx, err := getIndex(bucket, idxInfo.Type, fieldName)
			if err != nil {
				return err
			}

			err = idx.RemoveID(id)
			if err != nil {
				return err
			}
		}

		raw := bucket.Get(id)
		if raw == nil {
			return ErrNotFound
		}

		return bucket.Delete(id)
	})
}
