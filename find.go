package storm

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Find returns one or more records by the specified index
func (s *DB) Find(index string, value interface{}, to interface{}) error {
	ref := reflect.ValueOf(to)

	if ref.Kind() != reflect.Ptr || reflect.Indirect(ref).Kind() != reflect.Slice {
		return errors.New("provided target must be a pointer to a slice")
	}

	typ := reflect.Indirect(ref).Type().Elem()
	newElem := reflect.New(typ)

	d := structs.New(newElem.Interface())
	bucketName := d.Name()
	if bucketName == "" {
		return errors.New("provided target must have a name")
	}

	field, ok := d.FieldOk(index)
	if !ok {
		return fmt.Errorf("field %s not found", index)
	}

	tag := field.Tag("storm")
	if tag == "" {
		return fmt.Errorf("index %s not found", index)
	}

	return s.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		idx := bucket.Bucket([]byte(index))
		if idx == nil {
			return fmt.Errorf("index %s not found", index)
		}

		val, err := toBytes(value)
		if err != nil {
			return err
		}

		raw := idx.Get(val)
		if raw == nil {
			return errors.New("not found")
		}

		var list [][]byte

		if tag == "unique" {
			list = append(list, raw)
		} else if tag == "index" {
			err = json.Unmarshal(raw, &list)
			if err != nil {
				return err
			}

			if list == nil || len(list) == 0 {
				return errors.New("not found")
			}
		} else {
			return fmt.Errorf("unsupported struct tag %s", tag)
		}

		results := reflect.MakeSlice(reflect.Indirect(ref).Type(), len(list), len(list))

		for i := range list {
			raw = bucket.Get(list[i])
			if raw == nil {
				return errors.New("not found")
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
