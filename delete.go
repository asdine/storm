package storm

import "github.com/boltdb/bolt"

// Delete deletes a key from a bucket
func (n *Node) Delete(bucketName string, key interface{}) error {
	id, err := toBytes(key)
	if err != nil {
		return err
	}

	return n.s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket := n.getBucket(tx, bucketName)
		if bucket == nil {
			return ErrNotFound
		}

		return bucket.Delete(id)
	})
}

// Delete deletes a key from a bucket
func (s *DB) Delete(bucketName string, key interface{}) error {
	return s.root.Delete(bucketName, key)
}
