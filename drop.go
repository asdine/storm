package storm

import (
	"reflect"

	"github.com/boltdb/bolt"
)

// Drop a bucket
func (n *node) Drop(data interface{}) error {
	var bucketName string

	v := reflect.ValueOf(data)
	if v.Kind() != reflect.String {
		info, err := extract(&v)
		if err != nil {
			return err
		}

		bucketName = info.Name
	} else {
		bucketName = v.Interface().(string)
	}

	return n.readWriteTx(func(tx *bolt.Tx) error {
		return n.drop(tx, bucketName)
	})
}

func (n *node) drop(tx *bolt.Tx, bucketName string) error {
	bucket := n.GetBucket(tx)
	if bucket == nil {
		return tx.DeleteBucket([]byte(bucketName))
	}

	return bucket.DeleteBucket([]byte(bucketName))
}

// Drop a bucket
func (s *DB) Drop(data interface{}) error {
	return s.root.Drop(data)
}
