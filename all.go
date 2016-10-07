package storm

import (
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// AllByIndex gets all the records of a bucket that are indexed in the specified index
func (n *node) AllByIndex(fieldName string, to interface{}, options ...func(*index.Options)) error {
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

	cfg, err := extract(&newElem)
	if err != nil {
		return err
	}

	if cfg.ID.Name == fieldName {
		return n.All(to, options...)
	}

	opts := index.NewOptions()
	for _, fn := range options {
		fn(opts)
	}

	return n.readTx(func(tx *bolt.Tx) error {
		return n.allByIndex(tx, fieldName, cfg, &ref, opts)
	})
}

func (n *node) allByIndex(tx *bolt.Tx, fieldName string, cfg *structConfig, ref *reflect.Value, opts *index.Options) error {
	bucket := n.GetBucket(tx, cfg.Name)
	if bucket == nil {
		return ErrNotFound
	}

	idxInfo, ok := cfg.Fields[fieldName]
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

		err = n.s.codec.Unmarshal(raw, results.Index(i).Addr().Interface())
		if err != nil {
			return err
		}
	}

	reflect.Indirect(*ref).Set(results)
	return nil
}

// All gets all the records of a bucket.
// If there are no records it returns no error and the 'to' parameter is set to an empty slice.
func (n *node) All(to interface{}, options ...func(*index.Options)) error {
	opts := index.NewOptions()
	for _, fn := range options {
		fn(opts)
	}

	query := newQuery(n, q.True()).Limit(opts.Limit).Skip(opts.Skip)
	if opts.Reverse {
		query.Reverse()
	}

	err := query.Find(to)
	if err != nil && err != ErrNotFound {
		return err
	}

	if err == ErrNotFound {
		ref := reflect.ValueOf(to)
		results := reflect.MakeSlice(reflect.Indirect(ref).Type(), 0, 0)
		reflect.Indirect(ref).Set(results)
	}
	return nil
}

// AllByIndex gets all the records of a bucket that are indexed in the specified index
func (s *DB) AllByIndex(fieldName string, to interface{}, options ...func(*index.Options)) error {
	return s.root.AllByIndex(fieldName, to, options...)
}

// All get all the records of a bucket
func (s *DB) All(to interface{}, options ...func(*index.Options)) error {
	return s.root.All(to, options...)
}
