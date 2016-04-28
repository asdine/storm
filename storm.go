package storm

import (
	"os"
	"time"

	"github.com/asdine/storm/codec"
	"github.com/boltdb/bolt"
)

// Open opens a database at the given path with optional Storm options.
func Open(path string, stormOptions ...func(*DB) error) (*DB, error) {
	var err error

	s := &DB{
		Path:  path,
		Codec: defaultCodec,
	}

	for _, option := range stormOptions {
		option(s)
	}

	if s.boltMode == 0 {
		s.boltMode = 0600
	}

	if s.boltOptions == nil {
		s.boltOptions = &bolt.Options{Timeout: 1 * time.Second}
	}

	s.Bolt, err = bolt.Open(path, s.boltMode, s.boltOptions)
	if err != nil {
		return nil, err
	}

	s.root = &Node{s: s, rootBucket: s.rootBucket}

	return s, nil
}

// OpenWithOptions opens a database with the given boltDB options and optional Storm options.
// Deprecated: Use storm.Open with storm.BoltOptions instead.
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
	Codec codec.EncodeDecoder

	// Bolt is still easily accessible
	Bolt *bolt.DB

	// Bolt file mode
	boltMode os.FileMode

	// Bolt options
	boltOptions *bolt.Options

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

// WithTransaction returns a New Storm node that will use the given transaction.
func (s *DB) WithTransaction(tx *bolt.Tx) *Node {
	return s.root.WithTransaction(tx)
}
