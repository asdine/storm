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
	"github.com/coreos/bbolt"
	"github.com/stretchr/testify/require"
)

func TestNewStorm(t *testing.T) {
	db, err := Open("")

	require.Error(t, err)
	require.Nil(t, db)

	dir, err := ioutil.TempDir(os.TempDir(), "storm")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "storm.db")
	db, err = Open(file)
	defer db.Close()

	require.Implements(t, (*Node)(nil), db)

	require.NoError(t, err)
	require.NotNil(t, db.Bolt)
	require.Equal(t, defaultCodec, db.Codec())

	var v string
	err = db.Get(dbinfo, "version", &v)
	require.NoError(t, err)
	require.Equal(t, Version, v)
}

func TestNewStormWithStormOptions(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)

	dc := new(dummyCodec)
	db1, _ := Open(filepath.Join(dir, "storm1.db"), BoltOptions(0660, &bolt.Options{Timeout: 10 * time.Second}), Codec(dc), Root("a", "b"))
	require.Equal(t, dc, db1.Codec())
	require.Equal(t, []string{"a", "b"}, db1.Node.(*node).rootBucket)

	err := db1.Save(&SimpleUser{ID: 1})
	require.NoError(t, err)

	db2, _ := Open(filepath.Join(dir, "storm2.db"), Codec(dc))
	require.Equal(t, dc, db2.Codec())
}

func TestNewStormWithBatch(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)

	db1, _ := Open(filepath.Join(dir, "storm1.db"), Batch())
	defer db1.Close()

	require.True(t, db1.Node.(*node).batchMode)
	n := db1.From().(*node)
	require.True(t, n.batchMode)
	n = db1.WithBatch(true).(*node)
	require.True(t, n.batchMode)
	n = db1.WithBatch(false).(*node)
	require.False(t, n.batchMode)
	n = n.From().(*node)
	require.False(t, n.batchMode)
	n = n.WithBatch(true).(*node)
	require.True(t, n.batchMode)
}

func TestBoltDB(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	bDB, err := bolt.Open(filepath.Join(dir, "bolt.db"), 0600, &bolt.Options{Timeout: 10 * time.Second})
	require.NoError(t, err)
	// no need to close bolt.DB Storm will take care of it
	sDB, err := Open("my.db", UseDB(bDB))
	require.NoError(t, err)
	defer sDB.Close()
	err = sDB.Save(&SimpleUser{ID: 10})
	require.NoError(t, err)
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
	require.Nil(t, err)
	u2 := &SimpleUser{}
	err = defaultCodec.Unmarshal(encoded, u2)
	require.Nil(t, err)
	if !reflect.DeepEqual(u1, u2) {
		t.Fatal("Codec mismatch")
	}
}

func TestToBytes(t *testing.T) {
	b, err := toBytes([]byte("a slice of bytes"), nil)
	require.NoError(t, err)
	require.Equal(t, []byte("a slice of bytes"), b)

	b, err = toBytes("a string", nil)
	require.NoError(t, err)
	require.Equal(t, []byte("a string"), b)

	b, err = toBytes(&SimpleUser{ID: 10, Name: "John", age: 100}, json.Codec)
	require.NoError(t, err)
	require.Equal(t, `{"ID":10,"Name":"John"}`, string(b))

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

func createDB(t errorHandler, opts ...func(*Options) error) (*DB, func()) {
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
