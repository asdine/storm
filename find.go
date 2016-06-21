package storm

import (
	"fmt"
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/boltdb/bolt"
)

// Find returns one or more records by the specified index
func (n *Node) Find(fieldName string, value interface{}, to interface{}, options ...func(q *index.Options)) error {
	ref := reflect.ValueOf(to)

	if ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Slice {
		return ErrSlicePtrNeeded
	}

	typ := reflect.Indirect(ref).Type().Elem()

	bucketName := typ.Name()
	if bucketName == "" {
		return ErrNoName
	}

	field, ok := typ.FieldByName(fieldName)
	if !ok {
		return fmt.Errorf("field %s not found", fieldName)
	}

	tag := field.Tag.Get("storm")
	if tag == "" {
		return fmt.Errorf("index %s not found", fieldName)
	}

	val, err := toBytes(value, n.s.Codec)
	if err != nil {
		return err
	}

	opts := index.NewOptions()
	for _, fn := range options {
		fn(opts)
	}

	if n.tx != nil {
		return n.find(n.tx, bucketName, fieldName, tag, &ref, val, opts)
	}

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.find(tx, bucketName, fieldName, tag, &ref, val, opts)
	})
}

func (n *Node) find(tx *bolt.Tx, bucketName, fieldName, tag string, ref *reflect.Value, val []byte, opts *index.Options) error {
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

	results := reflect.MakeSlice(reflect.Indirect(*ref).Type(), len(list), len(list))

	for i := range list {
		raw := bucket.Get(list[i])
		if raw == nil {
			return ErrNotFound
		}

		err = n.s.Codec.Decode(raw, results.Index(i).Addr().Interface())
		if err != nil {
			return err
		}
	}

	reflect.Indirect(*ref).Set(results)
	return nil
}

// Find returns one or more records by the specified index
func (s *DB) Find(fieldName string, value interface{}, to interface{}, options ...func(q *index.Options)) error {
	return s.root.Find(fieldName, value, to, options...)
}
