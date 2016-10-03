package storm

import (
	"fmt"
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// Range returns one or more records by the specified index within the specified range
func (n *node) Range(fieldName string, min, max, to interface{}, options ...func(*index.Options)) error {
	sink, err := newListSink(n, to)
	if err != nil {
		return err
	}

	bucketName := sink.bucket()
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
		query := newQuery(n, q.And(q.Gte(fieldName, min), q.Lte(fieldName, max)))

		if opts.Reverse {
			query.Reverse()
		}

		err = n.readTx(func(tx *bolt.Tx) error {
			return query.query(tx, sink)
		})

		if err != nil {
			return err
		}

		return sink.flush()
	}

	mn, err := toBytes(min, n.s.codec)
	if err != nil {
		return err
	}

	mx, err := toBytes(max, n.s.codec)
	if err != nil {
		return err
	}

	if n.tx != nil {
		return n.rnge(n.tx, bucketName, fieldName, tag, sink, mn, mx, opts)
	}

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.rnge(tx, bucketName, fieldName, tag, sink, mn, mx, opts)
	})
}

func (n *node) rnge(tx *bolt.Tx, bucketName, fieldName, tag string, sink *listSink, min, max []byte, opts *index.Options) error {
	bucket := n.GetBucket(tx, bucketName)
	if bucket == nil {
		reflect.Indirect(sink.ref).SetLen(0)
		return nil
	}

	idx, err := getIndex(bucket, tag, fieldName)
	if err != nil {
		return err
	}

	list, err := idx.Range(min, max, opts)
	if err != nil {
		return err
	}

	sink.results = reflect.MakeSlice(reflect.Indirect(sink.ref).Type(), len(list), len(list))

	for i := range list {
		raw := bucket.Get(list[i])
		if raw == nil {
			return ErrNotFound
		}

		_, err = sink.filter(nil, bucket, list[i], raw)
		if err != nil {
			return err
		}
	}

	return sink.flush()
}

// Range returns one or more records by the specified index within the specified range
func (s *DB) Range(fieldName string, min, max, to interface{}, options ...func(*index.Options)) error {
	return s.root.Range(fieldName, min, max, to, options...)
}
