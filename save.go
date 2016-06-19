package storm

import (
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/boltdb/bolt"
)

// Save a structure
func (n *Node) Save(data interface{}) error {
	ref := reflect.ValueOf(data)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return ErrStructPtrNeeded
	}

	info, err := extract(&ref)
	if err != nil {
		return err
	}

	var id []byte

	if info.ID.IsZero {
		if !info.ID.IsOfIntegerFamily() || !n.s.autoIncrement {
			return ErrZeroID
		}
	} else {
		id, err = toBytes(info.ID.Value.Interface(), n.s.Codec)
		if err != nil {
			return err
		}
	}

	var raw []byte
	// postpone encoding if AutoIncrement mode if enabled
	if !n.s.autoIncrement {
		raw, err = n.s.Codec.Encode(data)
		if err != nil {
			return err
		}
	}

	if n.tx != nil {
		return n.save(n.tx, info, id, raw)
	}

	return n.s.Bolt.Update(func(tx *bolt.Tx) error {
		return n.save(tx, info, id, raw)
	})
}

func (n *Node) save(tx *bolt.Tx, info *modelInfo, id []byte, raw []byte) error {
	bucket, err := n.CreateBucketIfNotExists(tx, info.Name)
	if err != nil {
		return err
	}

	if info.ID.IsZero {
		// isZero and integer, generate next sequence
		intID, _ := bucket.NextSequence()

		// convert to the right integer size
		info.ID.Value.Set(reflect.ValueOf(intID).Convert(info.ID.Type()))
		id, err = toBytes(info.ID.Value.Interface(), n.s.Codec)
		if err != nil {
			return err
		}
	}

	if n.s.autoIncrement {
		raw, err = n.s.Codec.Encode(info.data)
		if err != nil {
			return err
		}
	}

	for fieldName, idxInfo := range info.Indexes {
		idx, err := getIndex(bucket, idxInfo.Type, fieldName)
		if err != nil {
			return err
		}

		err = idx.RemoveID(id)
		if err != nil {
			return err
		}

		if idxInfo.IsZero() {
			continue
		}

		value, err := toBytes(idxInfo.Value.Interface(), n.s.Codec)
		if err != nil {
			return err
		}

		err = idx.Add(value, id)
		if err != nil {
			if err == index.ErrAlreadyExists {
				return ErrAlreadyExists
			}
			return err
		}
	}

	return bucket.Put(id, raw)
}

// Save a structure
func (s *DB) Save(data interface{}) error {
	return s.root.Save(data)
}
