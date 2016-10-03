package storm

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/asdine/storm/codec/json"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	var v string
	err = db.Get(dbinfo, "version", &v)
	assert.NoError(t, err)
	assert.Equal(t, Version, v)
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

func TestNewStormWithBatch(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)

	db1, _ := Open(filepath.Join(dir, "storm1.db"), Batch())
	defer db1.Close()

	assert.True(t, db1.root.batchMode)
	n := db1.From().(*node)
	assert.True(t, n.batchMode)
	n = db1.WithBatch(true).(*node)
	assert.True(t, n.batchMode)
	n = db1.WithBatch(false).(*node)
	assert.False(t, n.batchMode)
	n = n.From().(*node)
	assert.False(t, n.batchMode)
	n = n.WithBatch(true).(*node)
	assert.True(t, n.batchMode)
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

func (c dummyCodec) Marshal(v interface{}) ([]byte, error) {
	return []byte("dummy"), nil
}

func (c dummyCodec) Unmarshal(b []byte, v interface{}) error {
	return nil
}

func (c dummyCodec) Name() string {
	return "dummy"
}

func TestCodec(t *testing.T) {
	u1 := &SimpleUser{Name: "John"}
	encoded, err := defaultCodec.Marshal(u1)
	assert.Nil(t, err)
	u2 := &SimpleUser{}
	err = defaultCodec.Unmarshal(encoded, u2)
	assert.Nil(t, err)
	if !reflect.DeepEqual(u1, u2) {
		t.Fatal("Codec mismatch")
	}
}

func TestToBytes(t *testing.T) {
	b, err := toBytes([]byte("a slice of bytes"), nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a slice of bytes"), b)

	b, err = toBytes("a string", nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("a string"), b)

	b, err = toBytes(&SimpleUser{ID: 10, Name: "John", age: 100}, json.Codec)
	assert.NoError(t, err)
	assert.Equal(t, `{"ID":10,"Name":"John"}`, string(b))

	tests := map[interface{}]interface{}{
		int(-math.MaxInt64):    int64(-math.MaxInt64),
		int(math.MaxInt64):     int64(math.MaxInt64),
		int8(-math.MaxInt8):    int8(-math.MaxInt8),
		int8(math.MaxInt8):     int8(math.MaxInt8),
		int16(-math.MaxInt16):  int16(-math.MaxInt16),
		int16(math.MaxInt16):   int16(math.MaxInt16),
		int32(-math.MaxInt32):  int32(-math.MaxInt32),
		int32(math.MaxInt32):   int32(math.MaxInt32),
		int64(-math.MaxInt64):  int64(-math.MaxInt64),
		int64(math.MaxInt64):   int64(math.MaxInt64),
		uint(math.MaxUint64):   uint64(math.MaxUint64),
		uint64(math.MaxUint64): uint64(math.MaxUint64),
	}

	for v, expected := range tests {
		b, err = toBytes(v, nil)
		require.NoError(t, err)
		require.NotNil(t, b)
		buf := bytes.NewReader(b)
		typ := reflect.TypeOf(expected)
		actual := reflect.New(typ)
		err = binary.Read(buf, binary.BigEndian, actual.Interface())
		require.NoError(t, err)
		require.Equal(t, expected, actual.Elem().Interface())
	}
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
