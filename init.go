package storm

import (
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/boltdb/bolt"
)

// Init creates the indexes and buckets for a given structure
func (n *Node) Init(data interface{}) error {
	v := reflect.ValueOf(data)
	info, err := extract(&v)
	if err != nil {
		return err
	}

	if n.tx != nil {
		return n.init(n.tx, info)
	}

	err = n.s.Bolt.Update(func(tx *bolt.Tx) error {
		return n.init(tx, info)
	})
	return err
}

func (n *Node) init(tx *bolt.Tx, info *modelInfo) error {
	bucket, err := n.CreateBucketIfNotExists(tx, info.Name)
	if err != nil {
		return err
	}

	for fieldName, idxInfo := range info.Indexes {
		switch idxInfo.Type {
		case tagUniqueIdx:
			_, err = index.NewUniqueIndex(bucket, []byte(indexPrefix+fieldName))
		case tagIdx:
			_, err = index.NewListIndex(bucket, []byte(indexPrefix+fieldName))
		default:
			err = ErrIdxNotFound
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// Init creates the indexes and buckets for a given structure
func (s *DB) Init(data interface{}) error {
	return s.root.Init(data)
}
