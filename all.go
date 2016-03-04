package storm

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// All get all the records of a bucket
func (s *DB) All(to interface{}) error {
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

	t, err := extractTags(newElem.Interface())
	if err != nil {
		return err
	}

	return s.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		idx := bucket.Bucket([]byte(t.IDField.Name()))
		if idx == nil {
			return fmt.Errorf("index %s not found", t.IDField.Name())
		}

		stats := idx.Stats()

		results := reflect.MakeSlice(reflect.Indirect(ref).Type(), stats.KeyN, stats.KeyN)

		i := 0
		err := idx.ForEach(func(k []byte, v []byte) error {
			raw := bucket.Get(v)
			err := json.Unmarshal(raw, results.Index(i).Addr().Interface())
			if err != nil {
				return err
			}
			i++
			return nil
		})
		if err != nil {
			return err
		}

		reflect.Indirect(ref).Set(results)
		return nil
	})
}
