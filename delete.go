package storm

import "github.com/boltdb/bolt"

// Delete deletes a key from a bucket
func (n *Node) Delete(bucketName string, key interface{}) error {
	id, err := toBytes(key, n.s.Codec)
	if err != nil {
		return err
	}

	if n.tx != nil {
		return n.delete(n.tx, bucketName, id)
	}

	return n.s.Bolt.Update(func(tx *bolt.Tx) error {
		return n.delete(tx, bucketName, id)
	})
}

func (n *Node) delete(tx *bolt.Tx, bucketName string, id []byte) error {
	bucket := n.GetBucket(tx, bucketName)
	if bucket == nil {
		return ErrNotFound
	}

	return bucket.Delete(id)
}

// Delete deletes a key from a bucket
func (s *DB) Delete(bucketName string, key interface{}) error {
	return s.root.Delete(bucketName, key)
}
