package storm

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Range returns one or more records by the specified index within the specified range
func (s *DB) Range(fieldName string, min []byte, max []byte, to interface{}) error {
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

	field, ok := d.FieldOk(fieldName)
	if !ok {
		return fmt.Errorf("field %s not found", fieldName)
	}

	tag := field.Tag("storm")
	if tag == "" {
		return fmt.Errorf("index %s not found", fieldName)
	}

	return s.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		idx, err := getIndex(bucket, tag, fieldName)
		if err != nil {
			if err == ErrIndexNotFound {
				return ErrNotFound
			}
			return err
		}

		mn, err := toBytes(min)
		if err != nil {
			return err
		}
		mx, err := toBytes(max)
		if err != nil {
			return err
		}

		list, err := idx.Range(mn, mx)
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
