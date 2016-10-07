package storm

import (
	"bytes"
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/boltdb/bolt"
)

// Save a structure
func (n *node) Save(data interface{}) error {
	ref := reflect.ValueOf(data)

	if !ref.IsValid() || ref.Kind() != reflect.Ptr || ref.Elem().Kind() != reflect.Struct {
		return ErrStructPtrNeeded
	}

	cfg, err := extract(&ref)
	if err != nil {
		return err
	}

	var id []byte

	if cfg.ID.IsZero {
		if !cfg.ID.IsInteger || !n.s.autoIncrement {
			return ErrZeroID
		}
	} else {
		id, err = toBytes(cfg.ID.Value.Interface(), n.s.codec)
		if err != nil {
			return err
		}
	}

	var raw []byte
	// postpone encoding if AutoIncrement mode if enabled
	if !n.s.autoIncrement {
		raw, err = n.s.codec.Marshal(data)
		if err != nil {
			return err
		}
	}

	return n.readWriteTx(func(tx *bolt.Tx) error {
		return n.save(tx, cfg, id, raw, data)
	})
}

func (n *node) save(tx *bolt.Tx, cfg *structConfig, id []byte, raw []byte, data interface{}) error {
	bucket, err := n.CreateBucketIfNotExists(tx, cfg.Name)
	if err != nil {
		return err
	}

	// save node configuration in the bucket
	err = n.saveMetadata(bucket)
	if err != nil {
		return err
	}

	if cfg.ID.IsZero {
		// isZero and integer, generate next sequence
		intID, _ := bucket.NextSequence()

		// convert to the right integer size
		cfg.ID.Value.Set(reflect.ValueOf(intID).Convert(cfg.ID.Value.Type()))
		id, err = toBytes(cfg.ID.Value.Interface(), n.s.codec)
		if err != nil {
			return err
		}
	}

	if data != nil {
		if n.s.autoIncrement {
			raw, err = n.s.codec.Marshal(data)
			if err != nil {
				return err
			}
		}
	}

	for fieldName, fieldCfg := range cfg.Fields {
		if fieldCfg.Index == "" {
			continue
		}

		idx, err := getIndex(bucket, fieldCfg.Index, fieldName)
		if err != nil {
			return err
		}

		if fieldCfg.IsZero {
			err = idx.RemoveID(id)
			if err != nil {
				return err
			}
			continue
		}

		value, err := toBytes(fieldCfg.Value.Interface(), n.s.codec)
		if err != nil {
			return err
		}

		var found bool
		idsSaved, err := idx.All(value, nil)
		if err != nil {
			return err
		}
		for _, idSaved := range idsSaved {
			if bytes.Compare(idSaved, id) == 0 {
				found = true
				break
			}
		}

		if found {
			continue
		}

		err = idx.RemoveID(id)
		if err != nil {
			return err
		}

		err = idx.Add(value, id)
		if err != nil {
			if err == index.ErrAlreadyExists {
				return ErrAlreadyExists
			}
			return err
		}
	}

	return bucket.Put(id, raw)
}

// Save a structure
func (s *DB) Save(data interface{}) error {
	return s.root.Save(data)
}
