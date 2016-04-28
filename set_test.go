package storm

import (
	"io/ioutil"
	"net/mail"
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/codec/gob"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	err := db.Set("b1", 10, 10)
	assert.NoError(t, err)
	err = db.Set("b1", "best friend's mail", &mail.Address{Name: "Gandalf", Address: "gandalf@lorien.ma"})
	assert.NoError(t, err)
	err = db.Set("b2", []byte("i'm already a slice of bytes"), "a value")
	assert.NoError(t, err)
	err = db.Set("b2", []byte("i'm already a slice of bytes"), nil)
	assert.NoError(t, err)
	err = db.Set("b1", 0, 100)
	assert.NoError(t, err)
	err = db.Set("b1", nil, 100)
	assert.Error(t, err)

	db.Bolt.View(func(tx *bolt.Tx) error {
		b1 := tx.Bucket([]byte("b1"))
		assert.NotNil(t, b1)
		b2 := tx.Bucket([]byte("b2"))
		assert.NotNil(t, b2)

		k1, err := toBytes(10, gob.Codec)
		assert.NoError(t, err)
		val := b1.Get(k1)
		assert.NotNil(t, val)

		k2 := []byte("best friend's mail")
		val = b1.Get(k2)
		assert.NotNil(t, val)

		k3, err := toBytes(0, gob.Codec)
		assert.NoError(t, err)
		val = b1.Get(k3)
		assert.NotNil(t, val)

		return nil
	})

	err = db.Set("", 0, 100)
	assert.Error(t, err)

	err = db.Set("b", nil, 100)
	assert.Error(t, err)

	err = db.Set("b", 10, nil)
	assert.NoError(t, err)

	err = db.Set("b", nil, nil)
	assert.Error(t, err)
}
