package storm

import (
	"fmt"
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
)

type item struct {
	value  *reflect.Value
	bucket *bolt.Bucket
	k      []byte
	v      []byte
}

func newSorter(node Node) *sorter {
	return &sorter{
		node:   node,
		rbTree: rbt.NewWithStringComparator(),
	}
}

// sorter is a filter
type sorter struct {
	node    Node
	rbTree  *rbt.Tree
	orderBy string
}

func (s *sorter) filter(r reflectSink, tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error) {
	newElem := r.elem()
	err := s.node.Codec().Unmarshal(v, newElem.Interface())
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
		if s.orderBy != "" {
			elm := reflect.Indirect(newElem).FieldByName(s.orderBy)
			if !elm.IsValid() {
				return false, fmt.Errorf("Unknown field %s", s.orderBy)
			}
			raw, err := toBytes(elm.Interface(), s.node.Codec())
			if err != nil {
				return false, err
			}
			s.rbTree.Put(string(raw), &item{
				bucket: bucket,
				value:  &newElem,
				k:      k,
				v:      v,
			})
			return false, nil
		}

		return r.add(bucket, k, v, newElem)
	}

	return false, nil
}

func (s *sorter) flush(snk reflectSink) error {
	s.orderBy = ""
	var err error
	var stop bool

	it := s.rbTree.Iterator()
	for it.Next() {
		item := it.Value().(*item)
		stop, err = snk.add(item.bucket, item.k, item.v, *item.value)
		if err != nil {
			return err
		}
		if stop {
			return snk.flush()
		}
	}

	return snk.flush()
}

type sink interface {
	filter(tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error)
	bucket() string
	flush() error
}

type reflectSink interface {
	elem() reflect.Value
	add(bucket *bolt.Bucket, k []byte, v []byte, elem reflect.Value) (bool, error)
	flush() error
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
		sorter:   newSorter(node),
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
	sorter   *sorter
}

func (l *listSink) filter(tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error) {
	return l.sorter.filter(l, tree, bucket, k, v)
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
	if l.sorter.orderBy != "" {
		return l.sorter.flush(l)
	}

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
		node:   node,
		ref:    ref,
		sorter: newSorter(node),
	}, nil
}

type firstSink struct {
	node   Node
	ref    reflect.Value
	skip   int
	found  bool
	sorter *sorter
}

func (f *firstSink) filter(tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error) {
	return f.sorter.filter(f, tree, bucket, k, v)
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
	if f.sorter.orderBy != "" {
		return f.sorter.flush(f)
	}

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
		node:   node,
		ref:    ref,
		sorter: newSorter(node),
	}, nil
}

type deleteSink struct {
	node    Node
	ref     reflect.Value
	skip    int
	limit   int
	removed int
	sorter  *sorter
}

func (d *deleteSink) filter(tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error) {
	return d.sorter.filter(d, tree, bucket, k, v)
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

	for fieldName, fieldCfg := range info.Fields {
		if fieldCfg.Index == "" {
			continue
		}
		idx, err := getIndex(bucket, fieldCfg.Index, fieldName)
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
	if d.sorter.orderBy != "" {
		return d.sorter.flush(d)
	}

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
		node:   node,
		ref:    ref,
		sorter: newSorter(node),
	}, nil
}

type countSink struct {
	node    Node
	ref     reflect.Value
	skip    int
	limit   int
	counter int
	sorter  *sorter
}

func (c *countSink) filter(tree q.Matcher, bucket *bolt.Bucket, k, v []byte) (bool, error) {
	return c.sorter.filter(c, tree, bucket, k, v)
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
