package storm

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// One returns one record by the specified index
func (s *DB) One(fieldName string, value interface{}, to interface{}) error {
	ref := reflect.ValueOf(to)

	if !ref.IsValid() || (ref.Kind() != reflect.Ptr && structs.IsStruct(to)) {
		return ErrStructPtrNeeded
	}

	if fieldName == "" {
		return ErrNotFound
	}

	d := structs.New(to)
	bucketName := d.Name()
	if bucketName == "" {
		return ErrNoName
	}

	field := d.Field(fieldName)
	tag := field.Tag("storm")
	if tag == "" {
		return fmt.Errorf("index %s doesn't exist", fieldName)
	}

	return s.Bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s doesn't exist", bucketName)
		}

		var idx Index
		var err error
		switch tag {
		case "unique":
			idx, err = NewUniqueIndex(bucket, []byte(indexPrefix+fieldName))
		case "index":
			idx, err = NewListIndex(bucket, []byte(indexPrefix+fieldName))
		default:
			err = ErrBadIndexType
		}

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

		id := idx.Get(val)
		if id == nil {
			return ErrNotFound
		}

		raw := bucket.Get(id)
		if raw == nil {
			return ErrNotFound
		}

		return json.Unmarshal(raw, to)
	})
}
