package storm

import (
	"time"

	"github.com/boltdb/bolt"
)

// Open storm
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

// DB struct
type DB struct {
	Path string
	Bolt *bolt.DB
}
