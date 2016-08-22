package storm

import (
	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// A Node in Storm represents the API to a BoltDB bucket.
type Node interface {
	Save(data interface{}) error
	One(fieldName string, value interface{}, to interface{}) error
	Find(fieldName string, value interface{}, to interface{}, options ...func(q *index.Options)) error
	DeleteStruct(data interface{}) error
	Remove(data interface{}) error
	All(to interface{}, options ...func(*index.Options)) error
	AllByIndex(fieldName string, to interface{}, options ...func(*index.Options)) error
	Drop(data interface{}) error
	Init(data interface{}) error
	Count(data interface{}) (int, error)
	Range(fieldName string, min, max, to interface{}, options ...func(*index.Options)) error
	RangeScan(min, max string) []Node
	PrefixScan(prefix string) []Node
	Select(matchers ...q.Matcher) Query
	Get(bucketName string, key interface{}, to interface{}) error
	Set(bucketName string, key interface{}, value interface{}) error
	Delete(bucketName string, key interface{}) error
	From(addend ...string) Node
	Bucket() []string
	GetBucket(tx *bolt.Tx, children ...string) *bolt.Bucket
	CreateBucketIfNotExists(tx *bolt.Tx, bucket string) (*bolt.Bucket, error)
	WithTransaction(tx *bolt.Tx) Node
	Begin(writable bool) (Node, error)
	Rollback() error
	Commit() error
}

// A Node in Storm represents the API to a BoltDB bucket.
type node struct {
	s *DB

	// The root bucket. In the normal, simple case this will be empty.
	rootBucket []string

	// Transaction object. Nil if not in transaction
	tx *bolt.Tx
}

// From returns a new Storm node with a new bucket root below the current.
// All DB operations on the new node will be executed relative to this bucket.
func (n node) From(addend ...string) Node {
	n.rootBucket = append(n.rootBucket, addend...)
	return &n
}

// WithTransaction returns a New Storm node that will use the given transaction.
func (n node) WithTransaction(tx *bolt.Tx) Node {
	n.tx = tx
	return &n
}

// Bucket returns the bucket name as a slice from the root.
// In the normal, simple case this will be empty.
func (n *node) Bucket() []string {
	return n.rootBucket
}
