package storm

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// AllByIndex gets all the records of a bucket that are indexed in the specified index
func (s *DB) AllByIndex(fieldName string, to interface{}) error {
	ref := reflect.ValueOf(to)

	if ref.Kind() != reflect.Ptr || reflect.Indirect(ref).Kind() != reflect.Slice {
		return ErrSlicePtrNeeded
	}

	typ := reflect.Indirect(ref).Type().Elem()
	newElem := reflect.New(typ)

	d := structs.New(newElem.Interface())
	bucketName := d.Name()
	if bucketName == "" {
		return ErrNoName
	}

	var tag string
	if fieldName == "" {
		info, err := extract(newElem.Interface())
		if err != nil {
			return err
		}

		if info.ID == nil {
			return ErrNoID
		}

		fieldName = info.ID.Name()
		tag = "unique"
	} else {
		field, ok := d.FieldOk(fieldName)
		if !ok {
			return fmt.Errorf("field %s not found", fieldName)
		}

		tag = field.Tag("storm")
		if tag == "" {
			return fmt.Errorf("index %s not found", fieldName)
		}
	}

	return s.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		var idx Index
		var err error
		switch tag {
		case "unique":
			idx, err = NewUniqueIndex(bucket, []byte(fieldName))
		case "index":
			idx, err = NewListIndex(bucket, []byte(fieldName))
		default:
			err = ErrBadIndexType
		}

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
