package storm

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// AllByIndex gets all the records of a bucket that are indexed in the specified index
func (s *DB) AllByIndex(index string, to interface{}) error {
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

	if index == "" {
		if t.ID == nil {
			return errors.New("missing struct tag id or ID field")
		}
		index = t.IDField.Name()
	}

	kind := indexKind(index, t)
	if kind == "" {
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

		stats := idx.Stats()
		results := reflect.MakeSlice(reflect.Indirect(ref).Type(), stats.KeyN, stats.KeyN)

		i := 0
		err := idx.ForEach(func(k []byte, v []byte) error {
			if kind == "list" {
				var list [][]byte
				err = json.Unmarshal(v, &list)
				if err != nil {
					return err
				}

				if len(list) == 0 {
					return nil
				}

				v = list[0]
			}
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

// All get all the records of a bucket
func (s *DB) All(to interface{}) error {
	return s.AllByIndex("", to)
}
