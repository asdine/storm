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
		d.autoIncrement = true
	}
}

// Root used to set the root bucket. See also the From method.
func Root(root ...string) func(*DB) {
	return func(d *DB) {
		d.rootBucket = root
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

	s.root = &Node{s: s, rootBucket: s.rootBucket}

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

	s.root = &Node{s: s, rootBucket: s.rootBucket}

	return s, nil
}

// DB is the wrapper around BoltDB. It contains an instance of BoltDB and uses it to perform all the
// needed operations
type DB struct {
	// Path of the database file
	Path string

	// Handles encoding and decoding of objects
	Codec EncodeDecoder

	// Bolt is still easily accessible
	Bolt *bolt.DB

	// Enable auto increment on empty integer fields
	autoIncrement bool

	// The root node that points to the root bucket.
	root *Node

	// The root bucket name
	rootBucket []string
}

// From returns a new Storm node with a new bucket root.
// All DB operations on the new node will be executed relative to the given
// bucket.
func (s *DB) From(root ...string) *Node {
	newNode := *s.root
	newNode.rootBucket = root
	return &newNode
}
