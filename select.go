package storm

import "github.com/asdine/storm/q"

// Select a list of records that match a list of matchers. Doesn't use indexes.
func (n *Node) Select(matchers ...q.Matcher) *Query {
	tree := q.And(matchers...)
	return &Query{
		tree: tree,
		node: n,
	}
}

// Select a list of records that match a list of matchers. Doesn't use indexes.
func (s *DB) Select(matchers ...q.Matcher) *Query {
	return s.root.Select(matchers...)
}
