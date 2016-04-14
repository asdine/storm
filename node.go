package storm

// A Node in Storm represents the API to a BoltDB bucket.
type Node struct {
	s *DB

	// The root bucket. In the normal, simple case this will be empty.
	rootBucket []string
}

// From returns a new Storm node with a new bucket root below the current.
// All DB operations on the new node will be executed relative to this bucket.
func (n Node) From(addend ...string) *Node {
	n.rootBucket = append(n.rootBucket, addend...)
	return &n
}
