package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/asdine/storm/codec/gob"
	"github.com/asdine/storm/codec/json"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestNewStorm(t *testing.T) {
	db, err := Open("")

	assert.Error(t, err)
	assert.Nil(t, db)

	dir, err := ioutil.TempDir(os.TempDir(), "storm")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "storm.db")
	db, err = Open(file)
	defer db.Close()

	assert.Implements(t, (*Node)(nil), db)

	assert.NoError(t, err)
	assert.Equal(t, file, db.Path)
	assert.NotNil(t, db.Bolt)
	assert.Equal(t, defaultCodec, db.Codec())
}

func TestNewStormWithStormOptions(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)

	dc := new(dummyCodec)
	db1, _ := Open(filepath.Join(dir, "storm1.db"), BoltOptions(0660, &bolt.Options{Timeout: 10 * time.Second}), Codec(dc), AutoIncrement(), Root("a", "b"))
	assert.Equal(t, dc, db1.Codec())
	assert.True(t, db1.autoIncrement)
	assert.Equal(t, os.FileMode(0660), db1.boltMode)
	assert.Equal(t, 10*time.Second, db1.boltOptions.Timeout)
	assert.Equal(t, []string{"a", "b"}, db1.rootBucket)
	assert.Equal(t, []string{"a", "b"}, db1.root.rootBucket)

	err := db1.Save(&SimpleUser{ID: 1})
	assert.NoError(t, err)

	db2, _ := Open(filepath.Join(dir, "storm2.db"), Codec(dc))
	assert.Equal(t, dc, db2.Codec())
}

func TestBoltDB(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	bDB, err := bolt.Open(filepath.Join(dir, "bolt.db"), 0600, &bolt.Options{Timeout: 10 * time.Second})
	assert.NoError(t, err)
	// no need to close bolt.DB Storm will take care of it
	sDB, err := Open("my.db", UseDB(bDB))
	assert.NoError(t, err)
	defer sDB.Close()
	err = sDB.Save(&SimpleUser{ID: 10})
	assert.NoError(t, err)
}

type dummyCodec int

func (c dummyCodec) Encode(v interface{}) ([]byte, error) {
	return []byte("dummy"), nil
}

func (c dummyCodec) Decode(b []byte, v interface{}) error {
	return nil
}

func TestCodec(t *testing.T) {
	u1 := &SimpleUser{Name: "John"}
	encoded, err := defaultCodec.Encode(u1)
	assert.Nil(t, err)
	u2 := &SimpleUser{}
	err = defaultCodec.Decode(encoded, u2)
	assert.Nil(t, err)
	if !reflect.DeepEqual(u1, u2) {
		t.Fatal("Codec mismatch")
	}
}

func TestToBytes(t *testing.T) {
	b, err := toBytes([]byte("a slice of bytes"), gob.Codec)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a slice of bytes"), b)

	b, err = toBytes("a string", gob.Codec)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a string"), b)

	b, err = toBytes(5, gob.Codec)
	assert.NoError(t, err)
	assert.NotNil(t, b)

	b, err = toBytes([]byte("Hey"), gob.Codec)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Hey"), b)
}

func TestToBytesWithCodec(t *testing.T) {
	b, err := toBytes([]byte("a slice of bytes"), json.Codec)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a slice of bytes"), b)

	b, err = toBytes("a string", json.Codec)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a string"), b)

	b, err = toBytes(&SimpleUser{ID: 10, Name: "John", age: 100}, json.Codec)
	assert.NoError(t, err)
	assert.Equal(t, `{"ID":10,"Name":"John"}`, string(b))
}

func createDB(t errorHandler, opts ...func(*DB) error) (*DB, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "storm")
	if err != nil {
		t.Error(err)
	}
	db, err := Open(filepath.Join(dir, "storm.db"), opts...)
	if err != nil {
		t.Error(err)
	}

	return db, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

type errorHandler interface {
	Error(args ...interface{})
}
