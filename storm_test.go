package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

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

	assert.NoError(t, err)
	assert.Equal(t, file, db.Path)
	assert.NotNil(t, db.Bolt)
	assert.Equal(t, defaultCodec, db.Codec)
}

func TestNewStormWithOptions(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := OpenWithOptions(filepath.Join(dir, "storm.db"), 0600, nil)
	defer db.Close()

	err := db.Save(&SimpleUser{ID: 10})
	assert.NoError(t, err)
}

func TestNewStormWithStormOptions(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)

	dc := new(dummyCodec)
	db1, _ := Open(filepath.Join(dir, "storm1.db"), BoltOptions(0660, &bolt.Options{Timeout: 10 * time.Second}), Codec(dc), AutoIncrement(), Root("a", "b"))
	assert.Equal(t, dc, db1.Codec)
	assert.True(t, db1.autoIncrement)
	assert.Equal(t, os.FileMode(0660), db1.boltMode)
	assert.Equal(t, 10*time.Second, db1.boltOptions.Timeout)
	assert.Equal(t, []string{"a", "b"}, db1.rootBucket)
	assert.Equal(t, []string{"a", "b"}, db1.root.rootBucket)

	err := db1.Save(&SimpleUser{ID: 1})
	assert.NoError(t, err)

	db2, _ := Open(filepath.Join(dir, "storm2.db"), Codec(dc))
	assert.Equal(t, dc, db2.Codec)
}

type dummyCodec int

func (c dummyCodec) Encode(v interface{}) ([]byte, error) {
	return []byte("dummy"), nil
}

func (c dummyCodec) Decode(b []byte, v interface{}) error {
	return nil
}
