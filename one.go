package storm

import (
	"github.com/asdine/storm/index"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
)

// One returns one record by the specified index
func (n *Node) One(fieldName string, value interface{}, to interface{}) error {
	sink, err := newFirstSink(to)
	if err != nil {
		return err
	}

	if fieldName == "" {
		return ErrNotFound
	}

	info, err := extract(&sink.ref)
	if err != nil {
		return err
	}

	_, ok := info.Indexes[fieldName]
	if !ok {
		query := newQuery(n, q.StrictEq(fieldName, value))

		if n.tx != nil {
			err = query.query(n.tx, sink)
		} else {
			err = n.s.Bolt.View(func(tx *bolt.Tx) error {
				return query.query(tx, sink)
			})
		}

		if err != nil {
			return err
		}

		return sink.flush()
	}

	val, err := toBytes(value, n.s.Codec)
	if err != nil {
		return err
	}

	if n.tx != nil {
		return n.one(n.tx, fieldName, info, to, val, fieldName == info.ID.Field.Name)
	}

	return n.s.Bolt.View(func(tx *bolt.Tx) error {
		return n.one(tx, fieldName, info, to, val, fieldName == info.ID.Field.Name)
	})
}

func (n *Node) one(tx *bolt.Tx, fieldName string, info *modelInfo, to interface{}, val []byte, skipIndex bool) error {
	bucket := n.GetBucket(tx, info.Name)
	if bucket == nil {
		return ErrNotFound
	}

	var id []byte
	if !skipIndex {
		idxInfo, ok := info.Indexes[fieldName]
		if !ok {
			return ErrNotFound
		}

		idx, err := getIndex(bucket, idxInfo.Type, fieldName)
		if err != nil {
			if err == index.ErrNotFound {
				return ErrNotFound
			}
			return err
		}

		id = idx.Get(val)
	} else {
		id = val
	}

	if id == nil {
		return ErrNotFound
	}

	raw := bucket.Get(id)
	if raw == nil {
		return ErrNotFound
	}

	return n.s.Codec.Decode(raw, to)
}

// One returns one record by the specified index
func (s *DB) One(fieldName string, value interface{}, to interface{}) error {
	return s.root.One(fieldName, value, to)
}
