package storm

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
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

// Save a structure
func (s *Storm) Save(data interface{}) error {
	if !structs.IsStruct(data) {
		return errors.New("provided data must be a struct or a pointer to struct")
	}

	t, err := extractTags(data)
	if err != nil {
		return err
	}

	if t.ID == nil {
		if t.IDField == nil {
			return errors.New("missing struct tag id")
		}
		t.ID = t.IDField
	}

	id, err := toBytes(t.ID)
	if err != nil {
		return err
	}

	err = s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(t.Name))
		if err != nil {
			return err
		}

		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return bucket.Put(id, raw)
	})
	return err
}
