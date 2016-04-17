package storm

import (
	"fmt"
	"reflect"

	"github.com/boltdb/bolt"
)

// AllByIndex gets all the records of a bucket that are indexed in the specified index
func (n *Node) AllByIndex(fieldName string, to interface{}) error {
	if fieldName == "" {
		return n.All(to)
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

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.allByIndex(tx, fieldName, info, &ref)
	})
}

func (n *Node) allByIndex(tx *bolt.Tx, fieldName string, info *modelInfo, ref *reflect.Value) error {
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
		if err == ErrIndexNotFound {
			return ErrNotFound
		}
		return err
	}

	list, err := idx.AllRecords()
	if err != nil {
		if err == ErrIndexNotFound {
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
func (n *Node) All(to interface{}) error {
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

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.all(tx, info, &ref, rtyp, typ)
	})
}

func (n *Node) all(tx *bolt.Tx, info *modelInfo, ref *reflect.Value, rtyp, typ reflect.Type) error {
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
func (s *DB) AllByIndex(fieldName string, to interface{}) error {
	return s.root.AllByIndex(fieldName, to)
}

// All get all the records of a bucket
func (s *DB) All(to interface{}) error {
	return s.root.All(to)
}
