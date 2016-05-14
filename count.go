package storm

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Count counts all the records of a bucket
func (n *Node) Count(data interface{}) (int, error) {
	if !structs.IsStruct(data) {
		return 0, ErrBadType
	}

	info, err := extract(data)
	if err != nil {
		return 0, err
	}

	var count int
	if n.tx != nil {
		err = n.count(n.tx, info, &count)
		return count, err
	}

	err = n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.count(tx, info, &count)
	})
	return count, err
}

func (n *Node) count(tx *bolt.Tx, info *modelInfo, count *int) error {
	bucket := n.GetBucket(tx, info.Name)
	if bucket == nil {
		return fmt.Errorf("bucket %s not found", info.Name)
	}

	*count = 0
	c := bucket.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if v == nil {
			continue
		}
		(*count)++
	}

	return nil
}

// Count counts all the records of a bucket
func (s *DB) Count(data interface{}) (int, error) {
	return s.root.Count(data)
}
