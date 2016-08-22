package storm

import (
	"fmt"
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// Find returns one or more records by the specified index
func (n *node) Find(fieldName string, value interface{}, to interface{}, options ...func(q *index.Options)) error {
	sink, err := newListSink(to)
	if err != nil {
		return err
	}

	bucketName := sink.name()
	if bucketName == "" {
		return ErrNoName
	}

	typ := reflect.Indirect(sink.ref).Type().Elem()

	field, ok := typ.FieldByName(fieldName)
	if !ok {
		return fmt.Errorf("field %s not found", fieldName)
	}

	opts := index.NewOptions()
	for _, fn := range options {
		fn(opts)
	}

	tag := field.Tag.Get("storm")
	if tag == "" {
		sink.limit = opts.Limit
		sink.skip = opts.Skip
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

	val, err := toBytes(value, n.s.Codec)
	if err != nil {
		return err
	}

	if n.tx != nil {
		return n.find(n.tx, bucketName, fieldName, tag, sink, val, opts)
	}

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.find(tx, bucketName, fieldName, tag, sink, val, opts)
	})
}

func (n *node) find(tx *bolt.Tx, bucketName, fieldName, tag string, sink *listSink, val []byte, opts *index.Options) error {
	bucket := n.GetBucket(tx, bucketName)
	if bucket == nil {
		return ErrNotFound
	}

	idx, err := getIndex(bucket, tag, fieldName)
	if err != nil {
		return err
	}

	list, err := idx.All(val, opts)
	if err != nil {
		if err == index.ErrNotFound {
			return ErrNotFound
		}
		return err
	}

	sink.results = reflect.MakeSlice(reflect.Indirect(sink.ref).Type(), len(list), len(list))

	for i := range list {
		raw := bucket.Get(list[i])
		if raw == nil {
			return ErrNotFound
		}

		elem := sink.elem()
		err = n.s.Codec.Decode(raw, elem.Interface())
		if err != nil {
			return err
		}

		_, err = sink.add(bucket, list[i], raw, elem)
		if err != nil {
			return err
		}
	}

	return sink.flush()
}

// Find returns one or more records by the specified index
func (s *DB) Find(fieldName string, value interface{}, to interface{}, options ...func(q *index.Options)) error {
	return s.root.Find(fieldName, value, to, options...)
}
