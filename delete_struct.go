package storm

import (
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/boltdb/bolt"
)

// DeleteStruct deletes a structure from the associated bucket
func (n *node) DeleteStruct(data interface{}) error {
	ref := reflect.ValueOf(data)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return ErrStructPtrNeeded
	}

	cfg, err := extract(&ref)
	if err != nil {
		return err
	}

	id, err := toBytes(cfg.ID.Value.Interface(), n.s.codec)
	if err != nil {
		return err
	}

	return n.readWriteTx(func(tx *bolt.Tx) error {
		return n.deleteStruct(tx, cfg, id)
	})
}

func (n *node) deleteStruct(tx *bolt.Tx, cfg *structConfig, id []byte) error {
	bucket := n.GetBucket(tx, cfg.Name)
	if bucket == nil {
		return ErrNotFound
	}

	for fieldName, fieldCfg := range cfg.Fields {
		if fieldCfg.Index == "" {
			continue
		}

		idx, err := getIndex(bucket, fieldCfg.Index, fieldName)
		if err != nil {
			return err
		}

		err = idx.RemoveID(id)
		if err != nil {
			if err == index.ErrNotFound {
				return ErrNotFound
			}
			return err
		}
	}

	raw := bucket.Get(id)
	if raw == nil {
		return ErrNotFound
	}

	return bucket.Delete(id)
}

// Remove deletes a structure from the associated bucket
// Deprecated: Use DeleteStruct instead.
func (n *node) Remove(data interface{}) error {
	return n.DeleteStruct(data)
}

// DeleteStruct deletes a structure from the associated bucket
func (s *DB) DeleteStruct(data interface{}) error {
	return s.root.DeleteStruct(data)
}

// Remove deletes a structure from the associated bucket
// Deprecated: Use DeleteStruct instead.
func (s *DB) Remove(data interface{}) error {
	return s.root.DeleteStruct(data)
}
