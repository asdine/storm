package storm

import (
	"reflect"

	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// Query is the low level query engine used by Storm. It allows to operate searches through an entire bucket.
type Query interface {
	// Skip matching records by the given number
	Skip(int) Query

	// Limit the results by the given number
	Limit(int) Query

	// Find a list of matching records
	Find(interface{}) error
}

type query struct {
	limit int
	skip  int
	tree  q.Matcher
	node  *Node
}

func (q *query) Skip(nb int) Query {
	q.skip = nb
	return q
}

func (q *query) Limit(nb int) Query {
	q.limit = nb
	return q
}

func (q *query) Find(to interface{}) error {
	ref := reflect.ValueOf(to)

	if ref.Kind() != reflect.Ptr || reflect.Indirect(ref).Kind() != reflect.Slice {
		return ErrSlicePtrNeeded
	}

	elemType := reflect.Indirect(ref).Type().Elem()

	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	newElem := reflect.New(elemType)

	info, err := extract(&newElem)
	if err != nil {
		return err
	}

	if q.node.tx != nil {
		return q.query(q.node.tx, info, &ref, elemType)
	}

	return q.node.s.Bolt.Update(func(tx *bolt.Tx) error {
		return q.query(tx, info, &ref, elemType)
	})
}

func (q *query) query(tx *bolt.Tx, info *modelInfo, ref *reflect.Value, elemType reflect.Type) error {
	results := reflect.MakeSlice(reflect.Indirect(*ref).Type(), 0, 0)
	bucket := q.node.GetBucket(tx, info.Name)

	realType := reflect.Indirect(*ref).Type().Elem()

	// we don't change state so queries can be replayed
	skip := q.skip

	if bucket != nil {
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				continue
			}

			newElem := reflect.New(elemType)
			err := q.node.s.Codec.Decode(v, newElem.Interface())
			if err != nil {
				return err
			}

			if q.tree.Match(newElem.Interface()) {

				if skip > 0 {
					skip--
					continue
				}

				if realType.Kind() == reflect.Ptr {
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
