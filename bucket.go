package storm

import (
	"github.com/boltdb/bolt"
)

func (n *Node) createBucketIfNotExists(tx *bolt.Tx, bucket string) (*bolt.Bucket, error) {
	var b *bolt.Bucket
	var err error

	bucketNames := append(n.rootBucket, bucket)

	for _, bucketName := range bucketNames {
		if b != nil {
			if b, err = b.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
				return nil, err
			}

		} else {
			if b, err = tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
				return nil, err
			}
		}
	}

	return b, nil
}

func (n *Node) getBucket(tx *bolt.Tx, bucket string) *bolt.Bucket {
	var b *bolt.Bucket

	bucketNames := append(n.rootBucket, bucket)

	for _, bucketName := range bucketNames {
		if b != nil {
			if b = b.Bucket([]byte(bucketName)); b == nil {
				return nil
			}
		} else {
			if b = tx.Bucket([]byte(bucketName)); b == nil {
				return nil
			}
		}
	}

	return b
}
