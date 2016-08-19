package storm

import (
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// AllByIndex gets all the records of a bucket that are indexed in the specified index
func (n *Node) AllByIndex(fieldName string, to interface{}, options ...func(*index.Options)) error {
	if fieldName == "" {
		return n.All(to, options...)
	}

	ref := reflect.ValueOf(to)

	if ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Slice {
		return ErrSlicePtrNeeded
	}

	typ := reflect.Indirect(ref).Type().Elem()

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	newElem := reflect.New(typ)

	info, err := extract(&newElem)
	if err != nil {
		return err
	}

	if info.ID.Field.Name == fieldName {
		return n.All(to, options...)
	}

	opts := index.NewOptions()
	for _, fn := range options {
		fn(opts)
	}

	if n.tx != nil {
		return n.allByIndex(n.tx, fieldName, info, &ref, opts)
	}

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.allByIndex(tx, fieldName, info, &ref, opts)
	})
}

func (n *Node) allByIndex(tx *bolt.Tx, fieldName string, info *modelInfo, ref *reflect.Value, opts *index.Options) error {
	bucket := n.GetBucket(tx, info.Name)
	if bucket == nil {
		return ErrNotFound
	}

	idxInfo, ok := info.Indexes[fieldName]
	if !ok {
		return ErrNotFound
	}

	idx, err := getIndex(bucket, idxInfo.Type, fieldName)
	if err != nil {
		return err
	}

	list, err := idx.AllRecords(opts)
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

// All gets all the records of a bucket
func (n *Node) All(to interface{}, options ...func(*index.Options)) error {
	sink, err := newListSink(to)
	if err != nil {
		return err
	}

	opts := index.NewOptions()
	for _, fn := range options {
		fn(opts)
	}

	sink.limit = opts.Limit
	sink.skip = opts.Skip

	query := newQuery(n, q.True())

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

// AllByIndex gets all the records of a bucket that are indexed in the specified index
func (s *DB) AllByIndex(fieldName string, to interface{}, options ...func(*index.Options)) error {
	return s.root.AllByIndex(fieldName, to, options...)
}

// All get all the records of a bucket
func (s *DB) All(to interface{}, options ...func(*index.Options)) error {
	return s.root.All(to, options...)
}
