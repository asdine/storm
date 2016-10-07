package storm

import (
	"bytes"
	"encoding/binary"
	"os"
	"time"

	"github.com/asdine/storm/codec"
	"github.com/asdine/storm/codec/json"
	"github.com/boltdb/bolt"
)

const (
	dbinfo         = "__storm_db"
	metadataBucket = "__storm_metadata"
)

// Defaults to json
var defaultCodec = json.Codec

// Open opens a database at the given path with optional Storm options.
func Open(path string, stormOptions ...func(*DB) error) (*DB, error) {
	var err error

	s := &DB{
		Path:  path,
		codec: defaultCodec,
	}

	for _, option := range stormOptions {
		if err = option(s); err != nil {
			return nil, err
		}
	}

	if s.boltMode == 0 {
		s.boltMode = 0600
	}

	if s.boltOptions == nil {
		s.boltOptions = &bolt.Options{Timeout: 1 * time.Second}
	}

	s.root = &node{s: s, rootBucket: s.rootBucket, codec: s.codec, batchMode: s.batchMode}

	// skip if UseDB option is used
	if s.Bolt == nil {
		s.Bolt, err = bolt.Open(path, s.boltMode, s.boltOptions)
		if err != nil {
			return nil, err
		}

		err = s.checkVersion()
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// DB is the wrapper around BoltDB. It contains an instance of BoltDB and uses it to perform all the
// needed operations
type DB struct {
	// Path of the database file
	Path string

	// Handles encoding and decoding of objects
	codec codec.MarshalUnmarshaler

	// Bolt is still easily accessible
	Bolt *bolt.DB

	// Bolt file mode
	boltMode os.FileMode

	// Bolt options
	boltOptions *bolt.Options

	// Enable auto increment on empty integer fields
	autoIncrement bool

	// The root node that points to the root bucket.
	root *node

	// The root bucket name
	rootBucket []string

	// Enable batch mode for read-write transaction, instead of update mode
	batchMode bool
}

// From returns a new Storm node with a new bucket root.
// All DB operations on the new node will be executed relative to the given
// bucket.
func (s *DB) From(root ...string) Node {
	newNode := *s.root
	newNode.rootBucket = root
	return &newNode
}

// WithTransaction returns a New Storm node that will use the given transaction.
func (s *DB) WithTransaction(tx *bolt.Tx) Node {
	return s.root.WithTransaction(tx)
}

// Bucket returns the root bucket name as a slice.
// In the normal, simple case this will be empty.
func (s *DB) Bucket() []string {
	return s.root.Bucket()
}

// Close the database
func (s *DB) Close() error {
	return s.Bolt.Close()
}

// Codec returns the EncodeDecoder used by this instance of Storm
func (s *DB) Codec() codec.MarshalUnmarshaler {
	return s.codec
}

// WithCodec returns a New Storm Node that will use the given Codec.
func (s *DB) WithCodec(codec codec.MarshalUnmarshaler) Node {
	n := s.From().(*node)
	n.codec = codec
	return n
}

// WithBatch returns a new Storm Node with the batch mode enabled.
func (s *DB) WithBatch(enabled bool) Node {
	n := s.From().(*node)
	n.batchMode = enabled
	return n
}

func (s *DB) checkVersion() error {
	var v string
	err := s.Get(dbinfo, "version", &v)
	if err != nil && err != ErrNotFound {
		return err
	}

	// for now, we only set the current version if it doesn't exist
	if v == "" {
		return s.Set(dbinfo, "version", Version)
	}

	return nil
}

// toBytes turns an interface into a slice of bytes
func toBytes(key interface{}, codec codec.MarshalUnmarshaler) ([]byte, error) {
	if key == nil {
		return nil, nil
	}
	switch t := key.(type) {
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	case int:
		return numbertob(int64(t))
	case uint:
		return numbertob(uint64(t))
	case int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return numbertob(t)
	default:
		return codec.Marshal(key)
	}
}

func numbertob(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func numberfromb(raw []byte) (int64, error) {
	r := bytes.NewReader(raw)
	var to int64
	err := binary.Read(r, binary.BigEndian, &to)
	if err != nil {
		return 0, err
	}
	return to, nil
}
