package storm

import (
	"reflect"

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

	var id []byte

	if info.ID.IsZero {
		if !info.ID.IsOfIntegerFamily() || !s.AutoIncrement {
			return ErrZeroID
		}
	} else {
		id, err = toBytes(info.ID.Value)
		if err != nil {
			return err
		}
	}

	return s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(info.Name))
		if err != nil {
			return err
		}

		if info.ID.IsZero {
			// isZero and integer, generate next sequence
			intID, _ := bucket.NextSequence()

			// convert to the right integer size
			err = info.ID.Field.Set(reflect.ValueOf(intID).Convert(info.ID.Type()).Interface())
			if err != nil {
				return err
			}

			id, err = toBytes(intID)
			if err != nil {
				return err
			}
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

		raw, err := s.Codec.Encode(data)
		if err != nil {
			return err
		}

		return bucket.Put(id, raw)
	})
}
