package index

import (
	"bytes"

	"github.com/AndersonBargas/rainstorm/v4/internal"
	bolt "go.etcd.io/bbolt"
)

// NewIDIndex loads a IDIndex
func NewIDIndex(parent *bolt.Bucket, indexName []byte) (*IDIndex, error) {
	return &IDIndex{
		IndexBucket: parent,
	}, nil
}

// IDIndex is an index that references unique values and the corresponding ID.
type IDIndex struct {
	IndexBucket *bolt.Bucket
}

// Add a value to the unique index
func (idx *IDIndex) Add(value []byte, targetID []byte) error {
	if value == nil || len(value) == 0 {
		return ErrNilParam
	}
	if targetID == nil || len(targetID) == 0 {
		return ErrNilParam
	}

	return nil
}

// Remove a value from the unique index
func (idx *IDIndex) Remove(value []byte) error {
	return nil
}

// RemoveID removes an ID from the unique index
func (idx *IDIndex) RemoveID(id []byte) error {
	return nil
}

// Get the id corresponding to the given value
func (idx *IDIndex) Get(value []byte) []byte {
	return idx.IndexBucket.Get(value)
}

// All returns all the ids corresponding to the given value
func (idx *IDIndex) All(value []byte, opts *Options) ([][]byte, error) {
	id := idx.IndexBucket.Get(value)
	if id != nil {
		return [][]byte{id}, nil
	}

	return nil, nil
}

// AllRecords returns all the IDs of this index
func (idx *IDIndex) AllRecords(opts *Options) ([][]byte, error) {
	var list [][]byte

	c := internal.Cursor{C: idx.IndexBucket.Cursor(), Reverse: opts != nil && opts.Reverse}

	for val, ident := c.First(); val != nil; val, ident = c.Next() {
		if opts != nil && opts.Skip > 0 {
			opts.Skip--
			continue
		}

		if opts != nil && opts.Limit == 0 {
			break
		}

		if opts != nil && opts.Limit > 0 {
			opts.Limit--
		}

		list = append(list, ident)
	}
	return list, nil
}

// Range returns the ids corresponding to the given range of values
func (idx *IDIndex) Range(min []byte, max []byte, opts *Options) ([][]byte, error) {
	var list [][]byte

	c := internal.RangeCursor{
		C:       idx.IndexBucket.Cursor(),
		Reverse: opts != nil && opts.Reverse,
		Min:     min,
		Max:     max,
		CompareFn: func(val, limit []byte) int {
			return bytes.Compare(val, limit)
		},
	}

	for ident, _ := c.First(); ident != nil && c.Continue(ident); ident, _ = c.Next() {
		if opts != nil && opts.Skip > 0 {
			opts.Skip--
			continue
		}

		if opts != nil && opts.Limit == 0 {
			break
		}

		if opts != nil && opts.Limit > 0 {
			opts.Limit--
		}

		list = append(list, ident)
	}
	return list, nil
}

// Prefix returns the ids whose values have the given prefix.
func (idx *IDIndex) Prefix(prefix []byte, opts *Options) ([][]byte, error) {
	var list [][]byte

	c := internal.PrefixCursor{
		C:       idx.IndexBucket.Cursor(),
		Reverse: opts != nil && opts.Reverse,
		Prefix:  prefix,
	}

	for ident, _ := c.First(); ident != nil && c.Continue(ident); ident, _ = c.Next() {
		if opts != nil && opts.Skip > 0 {
			opts.Skip--
			continue
		}

		if opts != nil && opts.Limit == 0 {
			break
		}

		if opts != nil && opts.Limit > 0 {
			opts.Limit--
		}

		list = append(list, ident)
	}
	return list, nil
}
