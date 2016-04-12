package storm

import (
	"os"
	"time"

	"github.com/boltdb/bolt"
)

// Codec used to set a custom encoder and decoder. The default is JSON.
func Codec(c EncodeDecoder) func(*DB) {
	return func(d *DB) {
		d.Codec = c
	}
}

// AutoIncrement used to enable bolt.NextSequence on empty integer ids.
func AutoIncrement() func(*DB) {
	return func(d *DB) {
		d.AutoIncrement = true
	}
}

// Open opens a database at the given path with optional Storm options.
func Open(path string, stormOptions ...func(*DB)) (*DB, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})

	if err != nil {
		return nil, err
	}

	s := &DB{
		Path:  path,
		Bolt:  db,
		Codec: defaultCodec,
	}

	for _, option := range stormOptions {
		option(s)
	}

	return s, nil
}

// OpenWithOptions opens a database with the given boltDB options and optional Storm options.
func OpenWithOptions(path string, mode os.FileMode, boltOptions *bolt.Options, stormOptions ...func(*DB)) (*DB, error) {
	db, err := bolt.Open(path, mode, boltOptions)

	if err != nil {
		return nil, err
	}

	s := &DB{
		Path:  path,
		Bolt:  db,
		Codec: defaultCodec,
	}

	for _, option := range stormOptions {
		option(s)
	}

	return s, nil
}

// DB is the wrapper around BoltDB. It contains an instance of BoltDB and uses it to perform all the
// needed operations
type DB struct {
	// Path of the database file
	Path string

	// Handles encoding and decoding of objects
	Codec EncodeDecoder

	// Enable auto increment on empty integer fields
	AutoIncrement bool

	// Bolt is still easily accessible
	Bolt *bolt.DB
}
