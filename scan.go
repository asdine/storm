package storm

import (
	"bytes"

	"github.com/boltdb/bolt"
)

// PrefixScan scans the root buckets for keys matching the given prefix.
func (s *DB) PrefixScan(prefix string) []*Node {
	return s.root.PrefixScan(prefix)
}

// PrefixScan scans the buckets in this node for keys matching the given prefix.
func (n *Node) PrefixScan(prefix string) []*Node {
	if n.tx != nil {
		return n.prefixScan(n.tx, prefix)
	}

	var nodes []*Node

	n.s.Bolt.View(func(tx *bolt.Tx) error {
		nodes = n.prefixScan(tx, prefix)
		return nil
	})

	return nodes
}

func (n *Node) prefixScan(tx *bolt.Tx, prefix string) []*Node {

	var (
		prefixBytes = []byte(prefix)
		nodes       []*Node
		c           = n.cursor(tx)
	)

	for k, _ := c.Seek(prefixBytes); bytes.HasPrefix(k, prefixBytes); k, _ = c.Next() {
		nodes = append(nodes, n.From(string(k)))
	}

	return nodes
}

// RangeScan scans the root buckets over a range such as a sortable time range.
func (s *DB) RangeScan(min, max string) []*Node {
	return s.root.RangeScan(min, max)
}

// RangeScan scans the buckets in this node  over a range such as a sortable time range.
func (n *Node) RangeScan(min, max string) []*Node {
	if n.tx != nil {
		return n.rangeScan(n.tx, min, max)
	}

	var nodes []*Node

	n.s.Bolt.View(func(tx *bolt.Tx) error {
		nodes = n.rangeScan(tx, min, max)
		return nil
	})

	return nodes
}

func (n *Node) rangeScan(tx *bolt.Tx, min, max string) []*Node {
	var (
		minBytes = []byte(min)
		maxBytes = []byte(max)
		nodes    []*Node
		c        = n.cursor(tx)
	)

	for k, _ := c.Seek(minBytes); k != nil && bytes.Compare(k, maxBytes) <= 0; k, _ = c.Next() {
		nodes = append(nodes, n.From(string(k)))
	}

	return nodes

}

func (n *Node) cursor(tx *bolt.Tx) *bolt.Cursor {

	var c *bolt.Cursor

	if len(n.rootBucket) > 0 {
		c = n.GetBucket(tx).Cursor()
	} else {
		c = tx.Cursor()
	}

	return c
}
