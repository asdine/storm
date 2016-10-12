package storm

import (
	"github.com/asdine/storm/internal"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// Query is the low level query engine used by Storm. It allows to operate searches through an entire bucket.
type Query interface {
	// Skip matching records by the given number
	Skip(int) Query

	// Limit the results by the given number
	Limit(int) Query

	// Order by the given field.
	OrderBy(string) Query

	// Reverse the order of the results
	Reverse() Query

	// Bucket specifies the bucket name
	Bucket(string) Query

	// Find a list of matching records
	Find(interface{}) error

	// First gets the first matching record
	First(interface{}) error

	// Delete all matching records
	Delete(interface{}) error

	// Count all the matching records
	Count(interface{}) (int, error)

	// Returns all the records without decoding them
	Raw() ([][]byte, error)

	// Execute the given function for each element
	RawEach(func([]byte, []byte) error) error
}

func newQuery(n *node, tree q.Matcher) *query {
	return &query{
		skip:   0,
		limit:  -1,
		node:   n,
		tree:   tree,
		sorter: newSorter(n),
	}
}

type query struct {
	limit   int
	skip    int
	reverse bool
	tree    q.Matcher
	node    *node
	bucket  string
	sorter  *sorter
}

func (q *query) Skip(nb int) Query {
	q.skip = nb
	return q
}

func (q *query) Limit(nb int) Query {
	q.limit = nb
	return q
}

func (q *query) OrderBy(field string) Query {
	q.sorter.orderBy = field
	return q
}

func (q *query) Reverse() Query {
	q.reverse = true
	q.sorter.reverse = true
	return q
}

func (q *query) Bucket(bucketName string) Query {
	q.bucket = bucketName
	return q
}

func (q *query) Find(to interface{}) error {
	sink, err := newListSink(q.node, to)
	if err != nil {
		return err
	}

	sink.limit = q.limit
	sink.skip = q.skip
	return q.runQuery(sink)
}

func (q *query) First(to interface{}) error {
	sink, err := newFirstSink(q.node, to)
	if err != nil {
		return err
	}

	sink.skip = q.skip
	return q.runQuery(sink)
}

func (q *query) Delete(kind interface{}) error {
	sink, err := newDeleteSink(q.node, kind)
	if err != nil {
		return err
	}

	sink.limit = q.limit
	sink.skip = q.skip

	return q.runQuery(sink)
}

func (q *query) Count(kind interface{}) (int, error) {
	sink, err := newCountSink(q.node, kind)
	if err != nil {
		return 0, err
	}

	sink.limit = q.limit
	sink.skip = q.skip

	err = q.runQuery(sink)
	if err != nil {
		return 0, err
	}

	return sink.counter, nil
}

func (q *query) Raw() ([][]byte, error) {
	sink := newRawSink()

	sink.limit = q.limit
	sink.skip = q.skip

	err := q.runQuery(sink)
	if err != nil {
		return nil, err
	}

	return sink.results, nil
}

func (q *query) RawEach(fn func([]byte, []byte) error) error {
	sink := newRawSink()

	sink.limit = q.limit
	sink.skip = q.skip
	sink.execFn = fn

	return q.runQuery(sink)
}

func (q *query) runQuery(sink sink) error {
	var err error

	if q.node.tx != nil {
		err = q.query(q.node.tx, sink)
	} else {
		err = q.node.s.Bolt.Update(func(tx *bolt.Tx) error {
			return q.query(tx, sink)
		})
	}

	return err
}

func (q *query) query(tx *bolt.Tx, sink sink) error {
	bucketName := q.bucket
	if bucketName == "" {
		bucketName = sink.bucketName()
	}
	bucket := q.node.GetBucket(tx, bucketName)

	if q.limit == 0 {
		return q.sorter.flush(sink)
	}

	if bucket != nil {
		c := internal.Cursor{C: bucket.Cursor(), Reverse: q.reverse}
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				continue
			}

			stop, err := q.sorter.filter(sink, q.tree, bucket, k, v)
			if err != nil {
				return err
			}

			if stop {
				break
			}
		}
	}

	return q.sorter.flush(sink)
}
