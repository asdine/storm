package storm

import (
	"github.com/asdine/storm/codec"
	"github.com/boltdb/bolt"
)

// A Node in Storm represents the API to a BoltDB bucket.
type Node interface {
	Tx
	TypeStore
	KeyValueStore
	BucketScanner
	From(addend ...string) Node
	Bucket() []string
	GetBucket(tx *bolt.Tx, children ...string) *bolt.Bucket
	CreateBucketIfNotExists(tx *bolt.Tx, bucket string) (*bolt.Bucket, error)
	WithTransaction(tx *bolt.Tx) Node
	Begin(writable bool) (Node, error)
	Codec() codec.EncodeDecoder
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

// Codec returns the EncodeDecoder used by this Node
func (n *node) Codec() codec.EncodeDecoder {
	return n.s.codec
}
