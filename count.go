package storm

import "github.com/asdine/storm/q"

// Count counts all the records of a bucket
func (n *node) Count(data interface{}) (int, error) {
	return n.Select(q.True()).Count(data)
}

// Count counts all the records of a bucket
func (s *DB) Count(data interface{}) (int, error) {
	return s.root.Count(data)
}
