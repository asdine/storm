package storm

import (
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// Query is the low level query engine used by Storm. It allows to operate searches through an entire bucket.
type Query interface {
	// Skip matching records by the given number
	Skip(int) Query

	// Limit the results by the given number
	Limit(int) Query

	// Reverse the order of the results
	Reverse() Query

	// Find a list of matching records
	Find(interface{}) error

	// First gets the first matching record
	First(interface{}) error

	// Delete all matching records
	Delete(interface{}) error

	// Count all the matching records
	Count(interface{}) (int, error)
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
	limit   int
	skip    int
	reverse bool
	tree    q.Matcher
	node    *Node
}

func (q *query) Skip(nb int) Query {
	q.skip = nb
	return q
}

func (q *query) Limit(nb int) Query {
	q.limit = nb
	return q
}

func (q *query) Reverse() Query {
	q.reverse = true
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

	return sink.flush()
}

func (q *query) First(to interface{}) error {
	sink, err := newFirstSink(to)
	if err != nil {
		return err
	}

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

	return sink.flush()
}

func (q *query) Delete(kind interface{}) error {
	sink, err := newDeleteSink(kind)
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

	return sink.flush()
}

func (q *query) Count(kind interface{}) (int, error) {
	sink, err := newCountSink(kind)
	if err != nil {
		return 0, err
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
		return 0, err
	}

	return sink.counter, sink.flush()
}

func (q *query) query(tx *bolt.Tx, sink sink) error {
	bucket := q.node.GetBucket(tx, sink.name())

	if q.limit == 0 {
		return nil
	}

	if bucket != nil {
		c := cursor{c: bucket.Cursor(), reverse: q.reverse}
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
				stop, err := sink.add(bucket, k, v, newElem)
				if stop || err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type cursor struct {
	c       *bolt.Cursor
	reverse bool
}

func (c *cursor) First() ([]byte, []byte) {
	if c.reverse {
		return c.c.Last()
	}

	return c.c.First()
}

func (c *cursor) Next() ([]byte, []byte) {
	if c.reverse {
		return c.c.Prev()
	}

	return c.c.Next()
}

type sink interface {
	elem() reflect.Value
	name() string
	add(bucket *bolt.Bucket, k []byte, v []byte, elem reflect.Value) (bool, error)
	flush() error
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

	if elemType.Name() == "" {
		return nil, ErrNoName
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
	idx      int
}

func (l *listSink) elem() reflect.Value {
	if l.results.IsValid() && l.idx < l.results.Len() {
		return l.results.Index(l.idx).Addr()
	}
	return reflect.New(l.elemType)
}

func (l *listSink) name() string {
	return l.elemType.Name()
}

func (l *listSink) add(bucket *bolt.Bucket, k []byte, v []byte, elem reflect.Value) (bool, error) {
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

	if l.idx == l.results.Len() {
		if l.isPtr {
			l.results = reflect.Append(l.results, elem)
		} else {
			l.results = reflect.Append(l.results, reflect.Indirect(elem))
		}
	}

	l.idx++

	return l.limit == 0, nil
}

func (l *listSink) flush() error {
	if l.results.IsValid() && l.results.Len() > 0 {
		reflect.Indirect(l.ref).Set(l.results)
		return nil
	}

	return ErrNotFound
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
	ref   reflect.Value
	skip  int
	found bool
}

func (f *firstSink) elem() reflect.Value {
	return reflect.New(reflect.Indirect(f.ref).Type())
}

func (f *firstSink) name() string {
	return reflect.Indirect(f.ref).Type().Name()
}

func (f *firstSink) add(bucket *bolt.Bucket, k []byte, v []byte, elem reflect.Value) (bool, error) {
	if f.skip > 0 {
		f.skip--
		return false, nil
	}

	reflect.Indirect(f.ref).Set(elem.Elem())
	f.found = true
	return true, nil
}

func (f *firstSink) flush() error {
	if !f.found {
		return ErrNotFound
	}

	return nil
}

func newDeleteSink(kind interface{}) (*deleteSink, error) {
	ref := reflect.ValueOf(kind)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return nil, ErrStructPtrNeeded
	}

	return &deleteSink{
		ref: ref,
	}, nil
}

type deleteSink struct {
	ref     reflect.Value
	skip    int
	limit   int
	removed int
}

func (d *deleteSink) elem() reflect.Value {
	return reflect.New(reflect.Indirect(d.ref).Type())
}

func (d *deleteSink) name() string {
	return reflect.Indirect(d.ref).Type().Name()
}

func (d *deleteSink) add(bucket *bolt.Bucket, k []byte, v []byte, elem reflect.Value) (bool, error) {
	if d.skip > 0 {
		d.skip--
		return false, nil
	}

	if d.limit > 0 {
		d.limit--
	}

	info, err := extract(&d.ref)
	if err != nil {
		return false, err
	}

	for fieldName, idxInfo := range info.Indexes {
		idx, err := getIndex(bucket, idxInfo.Type, fieldName)
		if err != nil {
			return false, err
		}

		err = idx.RemoveID(k)
		if err != nil {
			if err == index.ErrNotFound {
				return false, ErrNotFound
			}
			return false, err
		}
	}

	d.removed++
	return d.limit == 0, bucket.Delete(k)
}

func (d *deleteSink) flush() error {
	if d.removed == 0 {
		return ErrNotFound
	}

	return nil
}

func newCountSink(kind interface{}) (*countSink, error) {
	ref := reflect.ValueOf(kind)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return nil, ErrStructPtrNeeded
	}

	return &countSink{
		ref: ref,
	}, nil
}

type countSink struct {
	ref     reflect.Value
	skip    int
	limit   int
	counter int
}

func (c *countSink) elem() reflect.Value {
	return reflect.New(reflect.Indirect(c.ref).Type())
}

func (c *countSink) name() string {
	return reflect.Indirect(c.ref).Type().Name()
}

func (c *countSink) add(bucket *bolt.Bucket, k []byte, v []byte, elem reflect.Value) (bool, error) {
	if c.skip > 0 {
		c.skip--
		return false, nil
	}

	if c.limit > 0 {
		c.limit--
	}

	c.counter++
	return c.limit == 0, nil
}

func (c *countSink) flush() error {
	if c.counter == 0 {
		return ErrNotFound
	}

	return nil
}
