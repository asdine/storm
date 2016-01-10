package storm

import (
	"encoding/json"
	"errors"
	"strings"
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

	d := structs.New(data)
	fields := d.Fields()

	var id []byte
	var err error

	for _, f := range fields {
		if !f.IsExported() {
			continue
		}

		tag := f.Tag("storm")
		if tag == "id" {
			if f.IsZero() {
				return errors.New("id field must not be a zero value")
			}
			id, err = toBytes(f.Value())
			if err != nil {
				return err
			}
		}
	}

	if id == nil {
		return errors.New("missing struct tag id")
	}

	bucketName := strings.ToLower(d.Name())

	err = s.Bolt.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
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
