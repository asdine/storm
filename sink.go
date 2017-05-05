package storm

import (
	"reflect"
	"sort"

	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

type item struct {
	value  *reflect.Value
	bucket *bolt.Bucket
	k      []byte
	v      []byte
}

func newSorter(n Node, snk sink, tree q.Matcher, orderBy []string, reverse bool) *sorter {
	return &sorter{
		node:    n,
		sink:    snk,
		tree:    tree,
		orderBy: orderBy,
		reverse: reverse,
		list:    make([]*item, 0),
		err:     make(chan error),
		done:    make(chan struct{}),
	}
}

type sorter struct {
	node    Node
	sink    sink
	tree    q.Matcher
	list    []*item
	orderBy []string
	reverse bool
	err     chan error
	done    chan struct{}
}

func (s *sorter) filter(bucket *bolt.Bucket, k, v []byte) (bool, error) {
	rsink, ok := s.sink.(reflectSink)
	if !ok {
		return s.sink.add(&item{
			bucket: bucket,
			k:      k,
			v:      v,
		})
	}

	newElem := rsink.elem()
	if err := s.node.Codec().Unmarshal(v, newElem.Interface()); err != nil {
		return false, err
	}

	itm := &item{
		bucket: bucket,
		value:  &newElem,
		k:      k,
		v:      v,
	}

	if s.tree == nil {
		if len(s.orderBy) == 0 {
			return s.sink.add(itm)
		}
	} else {
		ok, err := s.tree.Match(newElem.Interface())
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	if len(s.orderBy) == 0 {
		return s.sink.add(itm)
	}

	s.list = append(s.list, itm)

	return false, nil
}

func (s *sorter) compareValue(left reflect.Value, right reflect.Value) int {
	if !left.IsValid() || !right.IsValid() {
		if left.IsValid() {
			return 1
		}
		return -1
	}

	switch left.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		l, r := left.Int(), right.Int()
		if l < r {
			return -1
		}
		if l > r {
			return 1
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		l, r := left.Uint(), right.Uint()
		if l < r {
			return -1
		}
		if l > r {
			return 1
		}
	case reflect.Float32, reflect.Float64:
		l, r := left.Float(), right.Float()
		if l < r {
			return -1
		}
		if l > r {
			return 1
		}
	case reflect.String:
		l, r := left.String(), right.String()
		if l < r {
			return -1
		}
		if l > r {
			return 1
		}
	default:
		rawLeft, err := toBytes(left.Interface(), s.node.Codec())
		if err != nil {
			return -1
		}
		rawRight, err := toBytes(right.Interface(), s.node.Codec())
		if err != nil {
			return 1
		}

		l, r := string(rawLeft), string(rawRight)
		if l < r {
			return -1
		}
		if l > r {
			return 1
		}
	}

	return 0
}

func (s *sorter) less(leftElem reflect.Value, rightElem reflect.Value) bool {
	for _, orderBy := range s.orderBy {
		leftField := reflect.Indirect(leftElem).FieldByName(orderBy)
		if !leftField.IsValid() {
			s.err <- ErrNotFound
			return false
		}
		rightField := reflect.Indirect(rightElem).FieldByName(orderBy)
		if !rightField.IsValid() {
			s.err <- ErrNotFound
			return false
		}

		direction := 1
		if s.reverse {
			direction = -1
		}

		switch s.compareValue(leftField, rightField) * direction {
		case -1:
			return true
		case 1:
			return false
		default:
			continue
		}
	}

	return false
}

func (s *sorter) flush() error {
	if len(s.orderBy) == 0 {
		return s.sink.flush()
	}

	go func() {
		sort.Sort(s)
		close(s.err)
	}()
	err := <-s.err
	close(s.done)

	if err != nil {
		return err
	}

	for _, itm := range s.list {
		if itm == nil {
			break
		}
		stop, err := s.sink.add(itm)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}

	return s.sink.flush()
}

func (s *sorter) Len() int {
	// skip if we encountered an earlier error
	select {
	case <-s.done:
		return 0
	default:
		return len(s.list)
	}
}

func (s *sorter) Swap(i, j int) {
	// skip if we encountered an earlier error
	select {
	case <-s.done:
		return
	default:
		s.list[i], s.list[j] = s.list[j], s.list[i]
	}
}

func (s *sorter) Less(i, j int) bool {
	// skip if we encountered an earlier error
	select {
	case <-s.done:
		return false
	default:
	}

	return s.less(*s.list[i].value, *s.list[j].value)
}

type sink interface {
	bucketName() string
	flush() error
	add(*item) (bool, error)
	readOnly() bool
}

type reflectSink interface {
	elem() reflect.Value
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
		results:  reflect.MakeSlice(reflect.Indirect(ref).Type(), 0, 0),
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

func (l *listSink) elem() reflect.Value {
	if l.results.IsValid() && l.idx < l.results.Len() {
		return l.results.Index(l.idx).Addr()
	}
	return reflect.New(l.elemType)
}

func (l *listSink) bucketName() string {
	return l.name
}

func (l *listSink) add(i *item) (bool, error) {
	if l.limit == 0 {
		return true, nil
	}

	if l.skip > 0 {
		l.skip--
		return false, nil
	}

	if l.limit > 0 {
		l.limit--
	}

	if l.idx == l.results.Len() {
		if l.isPtr {
			l.results = reflect.Append(l.results, *i.value)
		} else {
			l.results = reflect.Append(l.results, reflect.Indirect(*i.value))
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

func (l *listSink) readOnly() bool {
	return true
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

func (f *firstSink) elem() reflect.Value {
	return reflect.New(reflect.Indirect(f.ref).Type())
}

func (f *firstSink) bucketName() string {
	return reflect.Indirect(f.ref).Type().Name()
}

func (f *firstSink) add(i *item) (bool, error) {
	if f.skip > 0 {
		f.skip--
		return false, nil
	}

	reflect.Indirect(f.ref).Set(i.value.Elem())
	f.found = true
	return true, nil
}

func (f *firstSink) flush() error {
	if !f.found {
		return ErrNotFound
	}

	return nil
}

func (f *firstSink) readOnly() bool {
	return true
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

func (d *deleteSink) elem() reflect.Value {
	return reflect.New(reflect.Indirect(d.ref).Type())
}

func (d *deleteSink) bucketName() string {
	return reflect.Indirect(d.ref).Type().Name()
}

func (d *deleteSink) add(i *item) (bool, error) {
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
		idx, err := getIndex(i.bucket, fieldCfg.Index, fieldName)
		if err != nil {
			return false, err
		}

		err = idx.RemoveID(i.k)
		if err != nil {
			if err == index.ErrNotFound {
				return false, ErrNotFound
			}
			return false, err
		}
	}

	d.removed++
	return d.limit == 0, i.bucket.Delete(i.k)
}

func (d *deleteSink) flush() error {
	if d.removed == 0 {
		return ErrNotFound
	}

	return nil
}

func (d *deleteSink) readOnly() bool {
	return false
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

func (c *countSink) elem() reflect.Value {
	return reflect.New(reflect.Indirect(c.ref).Type())
}

func (c *countSink) bucketName() string {
	return reflect.Indirect(c.ref).Type().Name()
}

func (c *countSink) add(i *item) (bool, error) {
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
	return nil
}

func (c *countSink) readOnly() bool {
	return true
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

func (r *rawSink) add(i *item) (bool, error) {
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
		err := r.execFn(i.k, i.v)
		if err != nil {
			return false, err
		}
	} else {
		r.results = append(r.results, i.v)
	}

	return r.limit == 0, nil
}

func (r *rawSink) bucketName() string {
	return ""
}

func (r *rawSink) flush() error {
	return nil
}

func (r *rawSink) readOnly() bool {
	return true
}

func newEachSink(to interface{}) (*eachSink, error) {
	ref := reflect.ValueOf(to)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return nil, ErrStructPtrNeeded
	}

	return &eachSink{
		ref: ref,
	}, nil
}

type eachSink struct {
	skip   int
	limit  int
	ref    reflect.Value
	execFn func(interface{}) error
}

func (e *eachSink) elem() reflect.Value {
	return reflect.New(reflect.Indirect(e.ref).Type())
}

func (e *eachSink) bucketName() string {
	return reflect.Indirect(e.ref).Type().Name()
}

func (e *eachSink) add(i *item) (bool, error) {
	if e.limit == 0 {
		return true, nil
	}

	if e.skip > 0 {
		e.skip--
		return false, nil
	}

	if e.limit > 0 {
		e.limit--
	}

	err := e.execFn(i.value.Interface())
	if err != nil {
		return false, err
	}

	return e.limit == 0, nil
}

func (e *eachSink) flush() error {
	return nil
}

func (e *eachSink) readOnly() bool {
	return true
}
