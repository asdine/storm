package storm

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"encoding"

	"github.com/asdine/storm/id"
	"github.com/boltdb/bolt"
)

const (
	metaCodec = "codec"
)

func newMeta(b *bolt.Bucket, n Node) (*meta, error) {
	m := b.Bucket([]byte(metadataBucket))
	idProvider := n.IDProvider()
	if m != nil {
		name := m.Get([]byte(metaCodec))
		if string(name) != n.Codec().Name() {
			return nil, ErrDifferentCodec
		}
		return &meta{
			node:       n,
			bucket:     m,
			idProvider: idProvider,
		}, nil
	}

	m, err := b.CreateBucket([]byte(metadataBucket))
	if err != nil {
		return nil, err
	}

	m.Put([]byte(metaCodec), []byte(n.Codec().Name()))
	return &meta{
		node:       n,
		bucket:     m,
		idProvider: idProvider,
	}, nil
}

type meta struct {
	node       Node
	bucket     *bolt.Bucket
	idProvider id.New
}

func defaultIDProvider(start interface{}) id.Provider {
	return func(last []byte) (interface{}, error) {
		var err error
		counter := start.(int64)
		if last != nil {
			counter, err = numberfromb(last)
			if err != nil {
				return nil, err
			}
			counter++
		}

		return counter, nil
	}
}

func (m *meta) increment(field *fieldConfig) error {
	var nextFn id.Provider
	if m.idProvider != nil {
		nextFn = m.idProvider(field.IncrementStart)
	} else {
		nextFn = defaultIDProvider(field.IncrementStart)
	}

	raw := m.bucket.Get([]byte(field.Name + "counter"))
	next, err := nextFn(raw)
	if err != nil {
		return err
	}

	var nextRaw []byte

	if bm, ok := next.(encoding.BinaryMarshaler); ok {
		nextRaw, err = bm.MarshalBinary()
		if err != nil {
			return err
		}
	} else {
		nextRaw, err = numbertob(next)
		if err != nil {
			return err
		}
	}

	err = m.bucket.Put([]byte(field.Name+"counter"), nextRaw)
	if err != nil {
		return err
	}

	field.Value.Set(reflect.ValueOf(next).Convert(field.Value.Type()))
	field.IsZero = false
	return nil
}
