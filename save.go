package storm

import (
	"reflect"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Save a structure
func (n *Node) Save(data interface{}) error {
	if !structs.IsStruct(data) {
		return ErrBadType
	}

	info, err := extract(data)
	if err != nil {
		return err
	}

	var id []byte

	if info.ID.IsZero {
		if !info.ID.IsOfIntegerFamily() || !n.s.autoIncrement {
			return ErrZeroID
		}
	} else {
		id, err = toBytes(info.ID.Value, n.s.Codec)
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
		err = info.ID.Field.Set(reflect.ValueOf(intID).Convert(info.ID.Type()).Interface())
		if err != nil {
			return err
		}

		id, err = toBytes(info.ID.Field.Value(), n.s.Codec)
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

		if idxInfo.Field.IsZero() {
			continue
		}

		value, err := toBytes(idxInfo.Field.Value(), n.s.Codec)
		if err != nil {
			return err
		}

		err = idx.Add(value, id)
		if err != nil {
			return err
		}
	}

	return bucket.Put(id, raw)
}

// Save a structure
func (s *DB) Save(data interface{}) error {
	return s.root.Save(data)
}
