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

	prefixBytes := []byte(prefix)

	var nodes []*Node

	var c *bolt.Cursor

	if len(n.rootBucket) > 0 {
		c = n.GetBucket(tx).Cursor()
	} else {
		c = tx.Cursor()
	}

	for k, _ := c.Seek(prefixBytes); bytes.HasPrefix(k, prefixBytes); k, _ = c.Next() {
		nodes = append(nodes, n.From(string(k)))
	}
	return nodes
}
