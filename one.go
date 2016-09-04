package storm

import (
	"fmt"
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// One returns one record by the specified index
func (n *node) One(fieldName string, value interface{}, to interface{}) error {
	sink, err := newFirstSink(to)
	if err != nil {
		return err
	}

	bucketName := sink.name()
	if bucketName == "" {
		return ErrNoName
	}

	if fieldName == "" {
		return ErrNotFound
	}

	typ := reflect.Indirect(sink.ref).Type()

	field, ok := typ.FieldByName(fieldName)
	if !ok {
		return fmt.Errorf("field %s not found", fieldName)
	}

	tag := field.Tag.Get("storm")
	if tag == "" && fieldName != "ID" {
		query := newQuery(n, q.StrictEq(fieldName, value))

		if n.tx != nil {
			err = query.query(n.tx, sink)
		} else {
			err = n.s.Bolt.View(func(tx *bolt.Tx) error {
				return query.query(tx, sink)
			})
		}

		if err != nil {
			return err
		}

		return sink.flush()
	}

	val, err := toBytes(value, n.s.codec)
	if err != nil {
		return err
	}

	if n.tx != nil {
		return n.one(n.tx, bucketName, fieldName, tag, to, val, fieldName == "ID" || tag == "id")
	}

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.one(tx, bucketName, fieldName, tag, to, val, fieldName == "ID" || tag == "id")
	})
}

func (n *node) one(tx *bolt.Tx, bucketName, fieldName, tag string, to interface{}, val []byte, skipIndex bool) error {
	bucket := n.GetBucket(tx, bucketName)
	if bucket == nil {
		return ErrNotFound
	}

	var id []byte
	if !skipIndex {
		idx, err := getIndex(bucket, tag, fieldName)
		if err != nil {
			if err == index.ErrNotFound {
				return ErrNotFound
			}
			return err
		}

		id = idx.Get(val)
	} else {
		id = val
	}

	if id == nil {
		return ErrNotFound
	}

	raw := bucket.Get(id)
	if raw == nil {
		return ErrNotFound
	}

	return n.s.codec.Decode(raw, to)
}

// One returns one record by the specified index
func (s *DB) One(fieldName string, value interface{}, to interface{}) error {
	return s.root.One(fieldName, value, to)
}
