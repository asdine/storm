package storm

import (
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
)

// AllByIndex gets all the records of a bucket that are indexed in the specified index
func (n *Node) AllByIndex(fieldName string, to interface{}, options ...func(*queryOptions)) error {
	if fieldName == "" {
		return n.All(to, options...)
	}

	ref := reflect.ValueOf(to)

	if ref.Kind() != reflect.Ptr || reflect.Indirect(ref).Kind() != reflect.Slice {
		return ErrSlicePtrNeeded
	}

	typ := reflect.Indirect(ref).Type().Elem()

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	newElem := reflect.New(typ)

	info, err := extract(newElem.Interface())
	if err != nil {
		return err
	}

	if info.ID.Field.Name() == fieldName {
		return n.All(to, options...)
	}

	opts := newQueryOptions()
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

func (n *Node) allByIndex(tx *bolt.Tx, fieldName string, info *modelInfo, ref *reflect.Value, opts *queryOptions) error {
	bucket := n.GetBucket(tx, info.Name)
	if bucket == nil {
		return fmt.Errorf("bucket %s not found", info.Name)
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
func (n *Node) All(to interface{}, options ...func(*queryOptions)) error {
	ref := reflect.ValueOf(to)

	if ref.Kind() != reflect.Ptr || reflect.Indirect(ref).Kind() != reflect.Slice {
		return ErrSlicePtrNeeded
	}

	rtyp := reflect.Indirect(ref).Type().Elem()
	typ := rtyp

	if rtyp.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	newElem := reflect.New(typ)

	info, err := extract(newElem.Interface())
	if err != nil {
		return err
	}

	opts := newQueryOptions()
	for _, fn := range options {
		fn(opts)
	}

	if n.tx != nil {
		return n.all(n.tx, info, &ref, rtyp, typ, opts)
	}

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.all(tx, info, &ref, rtyp, typ, opts)
	})
}

func (n *Node) all(tx *bolt.Tx, info *modelInfo, ref *reflect.Value, rtyp, typ reflect.Type, opts *queryOptions) error {
	bucket := n.GetBucket(tx, info.Name)
	if bucket == nil {
		return fmt.Errorf("bucket %s not found", info.Name)
	}

	results := reflect.MakeSlice(reflect.Indirect(*ref).Type(), 0, 0)
	c := bucket.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			continue
		}

		if opts != nil && opts.skip > 0 {
			opts.skip--
			continue
		}

		if opts != nil && opts.limit == 0 {
			break
		}

		if opts != nil && opts.limit > 0 {
			opts.limit--
		}

		newElem := reflect.New(typ)
		err := n.s.Codec.Decode(v, newElem.Interface())
		if err != nil {
			return err
		}

		if rtyp.Kind() == reflect.Ptr {
			results = reflect.Append(results, newElem)
		} else {
			results = reflect.Append(results, reflect.Indirect(newElem))
		}
	}

	reflect.Indirect(*ref).Set(results)
	return nil
}

// AllByIndex gets all the records of a bucket that are indexed in the specified index
func (s *DB) AllByIndex(fieldName string, to interface{}, options ...func(*queryOptions)) error {
	return s.root.AllByIndex(fieldName, to, options...)
}

// All get all the records of a bucket
func (s *DB) All(to interface{}, options ...func(*queryOptions)) error {
	return s.root.All(to, options...)
}
