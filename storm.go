package storm

import (
	"os"
	"time"

	"github.com/boltdb/bolt"
)

// Open opens a database at the given path
func Open(path string) (*DB, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})

	if err != nil {
		return nil, err
	}

	return &DB{
		Path: path,
		Bolt: db,
	}, nil
}

// OpenWithOptions opens a database with the given boltDB options
func OpenWithOptions(path string, mode os.FileMode, options *bolt.Options) (*DB, error) {
	db, err := bolt.Open(path, mode, options)

	if err != nil {
		return nil, err
	}

	return &DB{
		Path: path,
		Bolt: db,
	}, nil
}

// DB is the wrapper around BoltDB. It contains an instance of BoltDB and uses it to perform all the
// needed operations
type DB struct {
	// Path of the database file
	Path string

	// Bolt is still easily accessible
	Bolt *bolt.DB
}
