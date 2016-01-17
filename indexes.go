package storm

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

func (s *DB) addToUniqueIndex(index []byte, id []byte, key []byte, parent *bolt.Bucket) error {
	bucket, err := parent.CreateBucketIfNotExists(index)
	if err != nil {
		return err
	}

	exists := bucket.Get(key)
	if exists != nil {
		return errors.New("already exists")
	}

	return bucket.Put(key, id)
}

func (s *DB) addToListIndex(index []byte, id []byte, key []byte, parent *bolt.Bucket) error {
	bucket, err := parent.CreateBucketIfNotExists(index)
	if err != nil {
		return err
	}

	var list [][]byte

	raw := bucket.Get(key)
	if raw != nil {
		err = json.Unmarshal(raw, &list)
		if err != nil {
			return err
		}
	}

	list = append(list, id)
	raw, err = json.Marshal(&list)
	if err != nil {
		return err
	}

	return bucket.Put(key, raw)
}

func (s *DB) deleteOldIndexes(parent *bolt.Bucket, id []byte, indexes []*structs.Field, unique bool) error {
	raw := parent.Get(id)
	if raw == nil {
		return nil
	}

	var content map[string]interface{}
	err := json.Unmarshal(raw, &content)
	if err != nil {
		return err
	}

	for _, field := range indexes {
		bucket := parent.Bucket([]byte(field.Name()))
		if bucket == nil {
			continue
		}

		f, ok := content[field.Name()]
		if !ok || !reflect.ValueOf(f).IsValid() {
			continue
		}

		key, err := toBytes(f)
		if err != nil {
			return err
		}

		if !unique {
			raw = bucket.Get(key)
			if raw == nil {
				continue
			}

			var list [][]byte
			err := json.Unmarshal(raw, &list)
			if err != nil {
				return err
			}

			for i, ident := range list {
				if bytes.Equal(ident, id) {
					list = append(list[0:i], list[i+1:]...)
					break
				}
			}

			raw, err = json.Marshal(list)
			if err != nil {
				return err
			}

			err = bucket.Put(key, raw)
		} else {
			err = bucket.Delete(key)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
