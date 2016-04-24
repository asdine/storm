package storm

import "github.com/boltdb/bolt"

// A Node in Storm represents the API to a BoltDB bucket.
type Node struct {
	s *DB

	// The root bucket. In the normal, simple case this will be empty.
	rootBucket []string

	// Transaction object. Nil if not in transaction
	tx *bolt.Tx
}

// From returns a new Storm node with a new bucket root below the current.
// All DB operations on the new node will be executed relative to this bucket.
func (n Node) From(addend ...string) *Node {
	n.rootBucket = append(n.rootBucket, addend...)
	return &n
}

// WithTransaction returns a New Storm node that will use the given transaction.
func (n Node) WithTransaction(tx *bolt.Tx) *Node {
	n.tx = tx
	return &n
}
