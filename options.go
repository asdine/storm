package storm

import (
	"os"

	"github.com/asdine/storm/codec"
	"github.com/boltdb/bolt"
)

// BoltOptions used to pass options to BoltDB.
func BoltOptions(mode os.FileMode, options *bolt.Options) func(*DB) error {
	return func(d *DB) error {
		d.boltMode = mode
		d.boltOptions = options
		return nil
	}
}

// Codec used to set a custom encoder and decoder. The default is GOB.
func Codec(c codec.EncodeDecoder) func(*DB) error {
	return func(d *DB) error {
		d.Codec = c
		return nil
	}
}

// AutoIncrement used to enable bolt.NextSequence on empty integer ids.
func AutoIncrement() func(*DB) error {
	return func(d *DB) error {
		d.autoIncrement = true
		return nil
	}
}

// Root used to set the root bucket. See also the From method.
func Root(root ...string) func(*DB) error {
	return func(d *DB) error {
		d.rootBucket = root
		return nil
	}
}

func newQueryOptions() *queryOptions {
	return &queryOptions{
		limit: -1,
		skip:  0,
	}
}

// queryOptions used to limit the customize the queries
type queryOptions struct {
	limit int
	skip  int
}

// Limit sets the maximum number of records to return
func Limit(limit int) func(q *queryOptions) {
	return func(q *queryOptions) {
		q.limit = limit
	}
}

// Skip sets the number of records to skip
func Skip(offset int) func(q *queryOptions) {
	return func(q *queryOptions) {
		q.skip = offset
	}
}
