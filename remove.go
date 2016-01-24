package storm

import (
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Remove removes a structure from the associated bucket
func (s *DB) Remove(data interface{}) error {
	if !structs.IsStruct(data) {
		return errors.New("provided data must be a struct or a pointer to struct")
	}

	t, err := extractTags(data)
	if err != nil {
		return err
	}

	if t.ID == nil {
		if t.IDField == nil {
			return errors.New("missing struct tag id")
		}
		t.ID = t.IDField
	}

	id, err := toBytes(t.ID)
	if err != nil {
		return err
	}

	return s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(t.Name))
		if bucket == nil {
			return fmt.Errorf("bucket %s doesn't exist", t.Name)
		}

		if len(t.Uniques) > 0 {
			err = s.deleteOldIndexes(bucket, id, t.Uniques, true)
			if err != nil {
				return err
			}
		}

		if len(t.Indexes) > 0 {
			err = s.deleteOldIndexes(bucket, id, t.Indexes, false)
			if err != nil {
				return err
			}
		}

		raw := bucket.Get(id)
		if raw == nil {
			return errors.New("not found")
		}

		return bucket.Delete(id)
	})
}
