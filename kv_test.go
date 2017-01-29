package storm

import (
	"net/mail"
	"testing"
	"time"

	"github.com/asdine/storm/codec/gob"
	"github.com/asdine/storm/codec/json"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	err := db.Set("trash", 10, 100)
	assert.NoError(t, err)

	var nb int
	err = db.Get("trash", 10, &nb)
	assert.NoError(t, err)
	assert.Equal(t, 100, nb)

	tm := time.Now()
	err = db.Set("logs", tm, "I'm hungry")
	assert.NoError(t, err)

	var message string
	err = db.Get("logs", tm, &message)
	assert.NoError(t, err)
	assert.Equal(t, "I'm hungry", message)

	var hand int
	err = db.Get("wallet", "100 bucks", &hand)
	assert.Equal(t, ErrNotFound, err)

	err = db.Set("wallet", "10 bucks", 10)
	assert.NoError(t, err)

	err = db.Get("wallet", "100 bucks", &hand)
	assert.Equal(t, ErrNotFound, err)

	err = db.Get("logs", tm, nil)
	assert.Equal(t, ErrPtrNeeded, err)

	err = db.Get("", nil, nil)
	assert.Equal(t, ErrPtrNeeded, err)

	err = db.Get("", "100 bucks", &hand)
	assert.Equal(t, ErrNotFound, err)
}

func TestGetBytes(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	err := db.SetBytes("trash", "a", []byte("hi"))
	assert.NoError(t, err)

	val, err := db.GetBytes("trash", "a")
	assert.NoError(t, err)
	assert.Equal(t, []byte("hi"), val)

	_, err = db.GetBytes("trash", "b")
	assert.Equal(t, ErrNotFound, err)
}

func TestSet(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

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

		k1, err := toBytes(10, json.Codec)
		assert.NoError(t, err)
		val := b1.Get(k1)
		assert.NotNil(t, val)

		k2 := []byte("best friend's mail")
		val = b1.Get(k2)
		assert.NotNil(t, val)

		k3, err := toBytes(0, json.Codec)
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

func TestSetMetadata(t *testing.T) {
	db, cleanup := createDB(t, Batch())
	defer cleanup()

	w := User{ID: 10, Name: "John"}
	err := db.Set("User", 10, &w)
	require.NoError(t, err)
	n := db.WithCodec(gob.Codec)
	err = n.Set("User", 10, &w)
	require.Equal(t, ErrDifferentCodec, err)
}

func TestDelete(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	err := db.Set("files", "myfile.csv", "a,b,c,d")
	assert.NoError(t, err)
	err = db.Delete("files", "myfile.csv")
	assert.NoError(t, err)
	err = db.Delete("files", "myfile.csv")
	assert.NoError(t, err)
	err = db.Delete("i don't exist", "myfile.csv")
	assert.Equal(t, ErrNotFound, err)
	err = db.Delete("", nil)
	assert.Equal(t, ErrNotFound, err)
}
