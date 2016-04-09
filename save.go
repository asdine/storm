package storm

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Save a structure
func (s *DB) Save(data interface{}) error {
	if !structs.IsStruct(data) {
		return ErrBadType
	}

	info, err := extract(data)
	if err != nil {
		return err
	}

	if info.ID.IsZero() {
		return ErrZeroID
	}

	id, err := toBytes(info.ID.Value())
	if err != nil {
		return err
	}

	return s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(info.Name))
		if err != nil {
			return err
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

			if idxInfo.Field.IsZero() {
				continue
			}

			value, err := toBytes(idxInfo.Field.Value())
			if err != nil {
				return err
			}

			err = idx.Add(value, id)
			if err != nil {
				return err
			}
		}

		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return bucket.Put(id, raw)
	})
}
