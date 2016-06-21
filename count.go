package storm

import (
	"reflect"

	"github.com/boltdb/bolt"
)

// Count counts all the records of a bucket
func (n *Node) Count(data interface{}) (int, error) {
	ref := reflect.ValueOf(data)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return 0, ErrStructPtrNeeded
	}

	info, err := extract(&ref)
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
	*count = 0

	bucket := n.GetBucket(tx, info.Name)
	if bucket == nil {
		return nil
	}

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
