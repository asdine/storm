package storm

import (
	"time"

	"github.com/boltdb/bolt"
)

// New storm
func New(dbPath string) (*Storm, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})

	if err != nil {
		return nil, err
	}

	return &Storm{
		Path: dbPath,
		Bolt: db,
	}, nil
}

// Storm struct
type Storm struct {
	Path string
	Bolt *bolt.DB
}
