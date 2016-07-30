package storm

import (
	"reflect"

	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// Select a list of records that match a list of criterias. Doesn't use indexes.
func (n *Node) Select(to interface{}, criterias ...q.Criteria) error {
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

	info, err := extract(&newElem)
	if err != nil {
		return err
	}

	tree := q.And(criterias...)

	if n.tx != nil {
		return n.selector(n.tx, info, &ref, rtyp, typ, tree)
	}

	return n.s.Bolt.Update(func(tx *bolt.Tx) error {
		return n.selector(tx, info, &ref, rtyp, typ, tree)
	})
}

func (n *Node) selector(tx *bolt.Tx, info *modelInfo, ref *reflect.Value, rtyp, typ reflect.Type, tree q.Criteria) error {
	results := reflect.MakeSlice(reflect.Indirect(*ref).Type(), 0, 0)
	bucket := n.GetBucket(tx, info.Name)

	if bucket != nil {
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

			if tree.Exec(newElem.Interface()) {
				if rtyp.Kind() == reflect.Ptr {
					results = reflect.Append(results, newElem)
				} else {
					results = reflect.Append(results, reflect.Indirect(newElem))
				}
			}
		}
	}

	reflect.Indirect(*ref).Set(results)
	return nil
}

// Select a list of records that match a list of criterias. Doesn't use indexes.
func (s *DB) Select(to interface{}, criterias ...q.Criteria) error {
	return s.root.Select(to, criterias...)
}
