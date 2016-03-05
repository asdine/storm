package storm

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Save a structure
func (s *DB) Save(data interface{}) error {
	if !structs.IsStruct(data) {
		return errors.New("provided data must be a struct or a pointer to struct")
	}

	t, err := extractTags(data)
	if err != nil {
		return err
	}

	if t.ZeroID {
		return errors.New("id field must not be a zero value")
	}

	if t.ID == nil {
		return errors.New("missing struct tag id or ID field")
	}

	id, err := toBytes(t.ID)
	if err != nil {
		return err
	}

	err = s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(t.Name))
		if err != nil {
			return err
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

		if t.Uniques != nil {
			for _, field := range t.Uniques {
				key, err := toBytes(field.Value())
				if err != nil {
					return err
				}

				err = s.addToUniqueIndex([]byte(field.Name()), id, key, bucket)
				if err != nil {
					return err
				}
			}
		}

		if t.Indexes != nil {
			for _, field := range t.Indexes {
				key, err := toBytes(field.Value())
				if err != nil {
					return err
				}

				err = s.addToListIndex([]byte(field.Name()), id, key, bucket)
				if err != nil {
					return err
				}
			}
		}

		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return bucket.Put(id, raw)
	})
	return err
}
