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

	// First gets the first matching record
	First(interface{}) error
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
	sink, err := newListSink(to)
	if err != nil {
		return err
	}

	sink.limit = q.limit
	sink.skip = q.skip

	if q.node.tx != nil {
		err = q.query(q.node.tx, sink)
	} else {
		err = q.node.s.Bolt.Update(func(tx *bolt.Tx) error {
			return q.query(tx, sink)
		})
	}

	if err != nil {
		return err
	}

	sink.flush()
	return nil
}

func (q *query) First(to interface{}) error {
	sink, err := newFirstSink(to)
	if err != nil {
		return err
	}

	sink.skip = q.skip

	if q.node.tx != nil {
		return q.query(q.node.tx, sink)
	}

	return q.node.s.Bolt.Update(func(tx *bolt.Tx) error {
		return q.query(tx, sink)
	})
}

func (q *query) query(tx *bolt.Tx, sink sink) error {
	bucket := q.node.GetBucket(tx, sink.name())

	if q.limit == 0 {
		return nil
	}

	if bucket != nil {
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				continue
			}

			newElem := sink.elem()
			err := q.node.s.Codec.Decode(v, newElem.Interface())
			if err != nil {
				return err
			}

			if q.tree.Match(newElem.Interface()) {
				stop, err := sink.add(newElem)
				if stop || err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type sink interface {
	elem() reflect.Value
	name() string
	add(elem reflect.Value) (bool, error)
	flush()
}

func newListSink(to interface{}) (*listSink, error) {
	ref := reflect.ValueOf(to)

	if ref.Kind() != reflect.Ptr || reflect.Indirect(ref).Kind() != reflect.Slice {
		return nil, ErrSlicePtrNeeded
	}

	sliceType := reflect.Indirect(ref).Type()
	elemType := sliceType.Elem()

	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	return &listSink{
		ref:      ref,
		isPtr:    sliceType.Elem().Kind() == reflect.Ptr,
		elemType: elemType,
	}, nil
}

type listSink struct {
	ref      reflect.Value
	results  reflect.Value
	elemType reflect.Type
	isPtr    bool
	skip     int
	limit    int
}

func (l *listSink) elem() reflect.Value {
	return reflect.New(l.elemType)
}

func (l *listSink) name() string {
	return l.elemType.Name()
}

func (l *listSink) add(elem reflect.Value) (bool, error) {
	if l.skip > 0 {
		l.skip--
		return false, nil
	}

	if !l.results.IsValid() {
		l.results = reflect.MakeSlice(reflect.Indirect(l.ref).Type(), 0, 0)
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

func (l *listSink) flush() {
	if l.results.IsValid() {
		reflect.Indirect(l.ref).Set(l.results)
	}
}

func newFirstSink(to interface{}) (*firstSink, error) {
	ref := reflect.ValueOf(to)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return nil, ErrStructPtrNeeded
	}

	return &firstSink{
		ref: ref,
	}, nil
}

type firstSink struct {
	ref  reflect.Value
	skip int
}

func (f *firstSink) elem() reflect.Value {
	return reflect.New(reflect.Indirect(f.ref).Type())
}

func (f *firstSink) name() string {
	return reflect.Indirect(f.ref).Type().Name()
}

func (f *firstSink) add(elem reflect.Value) (bool, error) {
	if f.skip > 0 {
		f.skip--
		return false, nil
	}

	reflect.Indirect(f.ref).Set(elem.Elem())
	return true, nil
}

func (f *firstSink) flush() {}
