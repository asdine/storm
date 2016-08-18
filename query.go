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

func newQuery(n *Node, tree q.Matcher) *query {
	return &query{
		skip:  0,
		limit: -1,
		node:  n,
		tree:  tree,
	}
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
	var err error
	ref := reflect.ValueOf(to)

	if ref.Kind() != reflect.Ptr || reflect.Indirect(ref).Kind() != reflect.Slice {
		return ErrSlicePtrNeeded
	}

	elemType := reflect.Indirect(ref).Type().Elem()

	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	sink := listSink{
		results: reflect.MakeSlice(reflect.Indirect(ref).Type(), 0, 0),
		isPtr:   reflect.Indirect(ref).Type().Elem().Kind() == reflect.Ptr,
		limit:   q.limit,
		skip:    q.skip,
	}

	if q.node.tx != nil {
		err = q.query(q.node.tx, elemType, &sink)
	} else {
		err = q.node.s.Bolt.Update(func(tx *bolt.Tx) error {
			return q.query(tx, elemType, &sink)
		})
	}

	if err != nil {
		return err
	}

	reflect.Indirect(ref).Set(sink.results)
	return nil
}

func (q *query) query(tx *bolt.Tx, elemType reflect.Type, sink sink) error {
	bucket := q.node.GetBucket(tx, elemType.Name())

	if q.limit == 0 {
		return nil
	}

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
				stop, err := sink.add(newElem)
				if err != nil {
					return err
				}
				if stop {
					return nil
				}
			}
		}
	}

	return nil
}

type sink interface {
	add(elem reflect.Value) (bool, error)
}

type listSink struct {
	results reflect.Value
	isPtr   bool
	skip    int
	limit   int
}

func (l *listSink) add(elem reflect.Value) (bool, error) {
	if l.skip > 0 {
		l.skip--
		return false, nil
	}

	if l.limit > 0 {
		l.limit--
	}

	if l.isPtr {
		l.results = reflect.Append(l.results, elem)
	} else {
		l.results = reflect.Append(l.results, reflect.Indirect(elem))
	}

	return l.limit == 0, nil
}
