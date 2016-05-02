package storm

import "github.com/boltdb/bolt"

// Drop a bucket
func (n *Node) Drop(bucketName string) error {
	if n.tx != nil {
		return n.drop(n.tx, bucketName)
	}

	return n.s.Bolt.Update(func(tx *bolt.Tx) error {
		return n.drop(tx, bucketName)
	})
}

func (n *Node) drop(tx *bolt.Tx, bucketName string) error {
	bucket := n.GetBucket(tx)
	if bucket == nil {
		return tx.DeleteBucket([]byte(bucketName))
	}

	return bucket.DeleteBucket([]byte(bucketName))
}

// Drop a bucket
func (s *DB) Drop(bucketName string) error {
	return s.root.Drop(bucketName)
}
