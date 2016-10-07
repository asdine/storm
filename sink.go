package storm

import (
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

type sink interface {
	filter(tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error)
	bucket() string
	flush() error
}

type reflectSink interface {
	elem() reflect.Value
	add(bucket *bolt.Bucket, k []byte, v []byte, elem reflect.Value) (bool, error)
}

func filter(s reflectSink, node Node, tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error) {
	newElem := s.elem()
	err := node.Codec().Unmarshal(v, newElem.Interface())
	if err != nil {
		return false, err
	}

	ok := tree == nil
	if !ok {
		ok, err = tree.Match(newElem.Interface())
		if err != nil {
			return false, err
		}
	}

	if ok {
		return s.add(bucket, k, v, newElem)
	}

	return false, nil
}

func newListSink(node Node, to interface{}) (*listSink, error) {
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
		node:     node,
		ref:      ref,
		isPtr:    sliceType.Elem().Kind() == reflect.Ptr,
		elemType: elemType,
		name:     elemType.Name(),
		limit:    -1,
	}, nil
}

type listSink struct {
	node     Node
	ref      reflect.Value
	results  reflect.Value
	elemType reflect.Type
	name     string
	isPtr    bool
	skip     int
	limit    int
	idx      int
}

func (l *listSink) filter(tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error) {
	return filter(l, l.node, tree, bucket, k, v)
}

func (l *listSink) elem() reflect.Value {
	if l.results.IsValid() && l.idx < l.results.Len() {
		return l.results.Index(l.idx).Addr()
	}
	return reflect.New(l.elemType)
}

func (l *listSink) bucket() string {
	return l.name
}

func (l *listSink) add(bucket *bolt.Bucket, k []byte, v []byte, elem reflect.Value) (bool, error) {
	if l.limit == 0 {
		return true, nil
	}

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

func newFirstSink(node Node, to interface{}) (*firstSink, error) {
	ref := reflect.ValueOf(to)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return nil, ErrStructPtrNeeded
	}

	return &firstSink{
		node: node,
		ref:  ref,
	}, nil
}

type firstSink struct {
	node  Node
	ref   reflect.Value
	skip  int
	found bool
}

func (f *firstSink) filter(tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error) {
	return filter(f, f.node, tree, bucket, k, v)
}

func (f *firstSink) elem() reflect.Value {
	return reflect.New(reflect.Indirect(f.ref).Type())
}

func (f *firstSink) bucket() string {
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

func newDeleteSink(node Node, kind interface{}) (*deleteSink, error) {
	ref := reflect.ValueOf(kind)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return nil, ErrStructPtrNeeded
	}

	return &deleteSink{
		node: node,
		ref:  ref,
	}, nil
}

type deleteSink struct {
	node    Node
	ref     reflect.Value
	skip    int
	limit   int
	removed int
}

func (d *deleteSink) filter(tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error) {
	return filter(d, d.node, tree, bucket, k, v)
}

func (d *deleteSink) elem() reflect.Value {
	return reflect.New(reflect.Indirect(d.ref).Type())
}

func (d *deleteSink) bucket() string {
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

	for fieldName, idxInfo := range info.Fields {
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

func newCountSink(node Node, kind interface{}) (*countSink, error) {
	ref := reflect.ValueOf(kind)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return nil, ErrStructPtrNeeded
	}

	return &countSink{
		node: node,
		ref:  ref,
	}, nil
}

type countSink struct {
	node    Node
	ref     reflect.Value
	skip    int
	limit   int
	counter int
}

func (c *countSink) filter(tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error) {
	return filter(c, c.node, tree, bucket, k, v)
}

func (c *countSink) elem() reflect.Value {
	return reflect.New(reflect.Indirect(c.ref).Type())
}

func (c *countSink) bucket() string {
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

func newRawSink() *rawSink {
	return &rawSink{
		limit: -1,
	}
}

type rawSink struct {
	results [][]byte
	skip    int
	limit   int
	execFn  func([]byte, []byte) error
}

func (r *rawSink) filter(tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error) {
	if r.limit == 0 {
		return true, nil
	}

	if r.skip > 0 {
		r.skip--
		return false, nil
	}

	if r.limit > 0 {
		r.limit--
	}

	if r.execFn != nil {
		err := r.execFn(k, v)
		if err != nil {
			return false, err
		}
	} else {
		r.results = append(r.results, v)
	}

	return r.limit == 0, nil
}

func (r *rawSink) bucket() string {
	return ""
}

func (r *rawSink) flush() error {
	return nil
}
