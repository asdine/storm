package storm

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
)

// AllByIndex gets all the records of a bucket that are indexed in the specified index
func (s *DB) AllByIndex(fieldName string, to interface{}) error {
	ref := reflect.ValueOf(to)

	if ref.Kind() != reflect.Ptr || reflect.Indirect(ref).Kind() != reflect.Slice {
		return ErrSlicePtrNeeded
	}

	typ := reflect.Indirect(ref).Type().Elem()
	newElem := reflect.New(typ)

	info, err := extract(newElem.Interface())
	if err != nil {
		return err
	}

	if fieldName == "" {
		fieldName = info.ID.Name()
	}

	idxInfo, ok := info.Indexes[fieldName]
	if !ok {
		return ErrNotFound
	}

	return s.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(info.Name))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", info.Name)
		}

		idx, err := getIndex(bucket, idxInfo.Type, fieldName)
		if err != nil {
			if err == ErrIndexNotFound {
				return ErrNotFound
			}
			return err
		}

		list, err := idx.AllRecords()
		if err != nil {
			if err == ErrIndexNotFound {
				return ErrNotFound
			}
			return err
		}

		results := reflect.MakeSlice(reflect.Indirect(ref).Type(), len(list), len(list))

		for i := range list {
			raw := bucket.Get(list[i])
			if raw == nil {
				return ErrNotFound
			}

			err = json.Unmarshal(raw, results.Index(i).Addr().Interface())
			if err != nil {
				return err
			}
		}

		reflect.Indirect(ref).Set(results)
		return nil
	})
}

// All get all the records of a bucket
func (s *DB) All(to interface{}) error {
	return s.AllByIndex("", to)
}
