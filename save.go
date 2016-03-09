package storm

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

const indexPrefix = "__storm_index_"

// Save a structure
func (s *DB) Save(data interface{}) error {
	if !structs.IsStruct(data) {
		return ErrBadType
	}

	info, err := extract(data)
	if err != nil {
		return err
	}

	if info.ID == nil {
		return ErrNoID
	}

	if info.ID.IsZero() {
		return ErrZeroID
	}

	if info.Name == "" {
		return ErrNoName
	}

	id, err := toBytes(info.ID.Value())
	if err != nil {
		return err
	}

	err = s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(info.Name))
		if err != nil {
			return err
		}

		var idx Index
		for fieldName, idxInfo := range info.Indexes {
			switch idxInfo.Type {
			case "unique":
				idx, err = NewUniqueIndex(bucket, []byte(indexPrefix+fieldName))
			case "index":
				idx, err = NewListIndex(bucket, []byte(indexPrefix+fieldName))
			default:
				err = ErrBadIndexType
			}

			if err != nil {
				return err
			}

			err = idx.RemoveID(id)
			if err != nil {
				return err
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
	return err
}
