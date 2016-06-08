package protobuf

import (
	"testing"
	"io/ioutil"
	"os"
	"path/filepath"

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
	u := SimpleUser{Id: 1, Name: "John"}
	err := db.Save(&u)
	assert.NoError(t, err)
}
