package storm

import (
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/boltdb/bolt"
)

// Remove removes a structure from the associated bucket
func (n *Node) Remove(data interface{}) error {
	ref := reflect.ValueOf(data)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return ErrStructPtrNeeded
	}

	info, err := extract(&ref)
	if err != nil {
		return err
	}

	id, err := toBytes(info.ID.Value.Interface(), n.s.Codec)
	if err != nil {
		return err
	}

	if n.tx != nil {
		return n.remove(n.tx, info, id)
	}

	return n.s.Bolt.Update(func(tx *bolt.Tx) error {
		return n.remove(tx, info, id)
	})
}

func (n *Node) remove(tx *bolt.Tx, info *modelInfo, id []byte) error {
	bucket := n.GetBucket(tx, info.Name)
	if bucket == nil {
		return ErrNotFound
	}

	for fieldName, idxInfo := range info.Indexes {
		idx, err := getIndex(bucket, idxInfo.Type, fieldName)
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

// Remove removes a structure from the associated bucket
func (s *DB) Remove(data interface{}) error {
	return s.root.Remove(data)
}
