package storm

import (
	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
)

// Init creates the indexes and buckets for a given structure
func (n *Node) Init(data interface{}) error {
	if !structs.IsStruct(data) {
		return ErrBadType
	}

	info, err := extract(data)
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
			_, err = NewUniqueIndex(bucket, []byte(indexPrefix+fieldName))
		case tagIdx:
			_, err = NewListIndex(bucket, []byte(indexPrefix+fieldName))
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
