package storm

import (
	"fmt"
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// TypeStore stores user defined types in BoltDB
type TypeStore interface {
	Finder
	// Init creates the indexes and buckets for a given structure
	Init(data interface{}) error

	// Save a structure
	Save(data interface{}) error

	// Update a structure
	Update(data interface{}) error

	// UpdateField updates a single field
	UpdateField(data interface{}, fieldName string, value interface{}) error

	// Drop a bucket
	Drop(data interface{}) error

	// DeleteStruct deletes a structure from the associated bucket
	DeleteStruct(data interface{}) error

	// Remove deletes a structure from the associated bucket
	// Deprecated: Use DeleteStruct instead.
	Remove(data interface{}) error
}

// A Finder can fetch types from BoltDB
type Finder interface {
	// One returns one record by the specified index
	One(fieldName string, value interface{}, to interface{}) error

	// Find returns one or more records by the specified index
	Find(fieldName string, value interface{}, to interface{}, options ...func(q *index.Options)) error

	// AllByIndex gets all the records of a bucket that are indexed in the specified index
	AllByIndex(fieldName string, to interface{}, options ...func(*index.Options)) error

	// All gets all the records of a bucket.
	// If there are no records it returns no error and the 'to' parameter is set to an empty slice.
	All(to interface{}, options ...func(*index.Options)) error

	// Select a list of records that match a list of matchers. Doesn't use indexes.
	Select(matchers ...q.Matcher) Query

	// Range returns one or more records by the specified index within the specified range
	Range(fieldName string, min, max, to interface{}, options ...func(*index.Options)) error

	// Count counts all the records of a bucket
	Count(data interface{}) (int, error)
}

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

		if opts.Reverse {
			query.Reverse()
		}

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
		err = n.s.codec.Decode(raw, elem.Interface())
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
