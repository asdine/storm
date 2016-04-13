package storm

import (
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Find returns one or more records by the specified index
func (n *Node) Find(fieldName string, value interface{}, to interface{}) error {
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

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		bucket := n.getBucket(tx, bucketName)
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

		val, err := toBytes(value)
		if err != nil {
			return err
		}

		list, err := idx.All(val)
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

			err = n.s.Codec.Decode(raw, results.Index(i).Addr().Interface())
			if err != nil {
				return err
			}
		}

		reflect.Indirect(ref).Set(results)
		return nil
	})
}

// Find returns one or more records by the specified index
func (s *DB) Find(fieldName string, value interface{}, to interface{}) error {
	return s.root.Find(fieldName, value, to)
}
