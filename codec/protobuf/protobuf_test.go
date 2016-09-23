package protobuf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/internal"
	"github.com/stretchr/testify/assert"
)

func TestProtobuf(t *testing.T) {
	u1 := SimpleUser{Id: 1, Name: "John"}
	u2 := SimpleUser{}
	internal.RoundtripTester(t, Codec, &u1, &u2)
	assert.True(t, u1.Id == u2.Id)
}

func TestSave(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"), storm.Codec(Codec))
	u1 := SimpleUser{Id: 1, Name: "John"}
	err := db.Save(&u1)
	assert.NoError(t, err)
	u2 := SimpleUser{}
	err = db.One("Id", uint64(1), &u2)
	assert.NoError(t, err)
	assert.Equal(t, u2.Name, u1.Name)
}

func TestGetSet(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := storm.Open(filepath.Join(dir, "storm.db"), storm.Codec(Codec))
	err := db.Set("bucket", "key", "value")
	assert.NoError(t, err)
	var s string
	err = db.Get("bucket", "key", &s)
	assert.NoError(t, err)
	assert.Equal(t, "value", s)
}
