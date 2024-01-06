package index_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/AndersonBargas/rainstorm/v5"
	"github.com/AndersonBargas/rainstorm/v5/index"
	"github.com/AndersonBargas/rainstorm/v5/q"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

// SimpleLogin represents a simple login model
type SimpleLogin struct {
	Email    string `rainstorm:"id"`
	Password string
}

// SimpleProduct represents a simple product model
type SimpleProduct struct {
	Barcode     int `rainstorm:"id,increment"`
	Description string
}

func TestIDIndex(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "rainstorm")
	defer os.RemoveAll(dir)
	db, _ := rainstorm.Open(filepath.Join(dir, "rainstorm.db"))
	defer db.Close()

	err := db.Init(&SimpleLogin{})
	require.NoError(t, err)

	simpleLogin := &SimpleLogin{
		Email:    "unique@example.org",
		Password: "OoopsRAWpassword!",
	}

	err = db.Save(simpleLogin)
	require.NoError(t, err)

	var simpleLogins []SimpleLogin

	err = db.AllByIndex("Email", &simpleLogins)
	require.NoError(t, err)
	require.Len(t, simpleLogins, 1)

	err = db.Prefix("Email", "uni", &simpleLogins)
	require.NoError(t, err)

	err = db.Select(q.Eq("Email", "unique@example.org")).First(&SimpleLogin{})
	require.NoError(t, err)

	err = db.Set("loggedInUsers", 1, &simpleLogin)
	require.NoError(t, err)

	err = db.Delete("loggedInUsers", 1)
	require.NoError(t, err)
}

func TestIDIndexPrefix(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "rainstorm")
	defer os.RemoveAll(dir)
	db, _ := rainstorm.Open(filepath.Join(dir, "rainstorm.db"))
	defer db.Close()

	err := db.Init(&SimpleLogin{})
	require.NoError(t, err)

	simpleLogin := &SimpleLogin{
		Email:    "unique@example.org",
		Password: "OoopsRAWpassword!",
	}

	err = db.Save(simpleLogin)
	require.NoError(t, err)

	simpleLogin = &SimpleLogin{
		Email:    "unique123@example.org",
		Password: "OoopsRAWpasswordAgain!",
	}

	err = db.Save(simpleLogin)
	require.NoError(t, err)

	var simpleLogins []SimpleLogin

	err = db.Prefix("Email", "uni", &simpleLogins)
	require.NoError(t, err)

	setSkip := func(opt *index.Options) {
		opt.Skip = 1
	}
	err = db.Prefix("Email", "uni", &simpleLogins, setSkip)
	require.NoError(t, err)

	setZeroedLimit := func(opt *index.Options) {
		opt.Limit = 0
	}
	err = db.Prefix("Email", "uni", &simpleLogins, setZeroedLimit)
	require.Error(t, err)
	require.True(t, rainstorm.ErrNotFound == err)

	setLimit := func(opt *index.Options) {
		opt.Limit = 1
	}
	err = db.Prefix("Email", "uni", &simpleLogins, setLimit)
	require.NoError(t, err)
}

func TestIDIndexRange(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "rainstorm")
	defer os.RemoveAll(dir)
	db, _ := rainstorm.Open(filepath.Join(dir, "rainstorm.db"))
	defer db.Close()

	err := db.Init(&SimpleProduct{})
	require.NoError(t, err)

	for i := 1; i <= 50; i++ {
		simpleProduct := &SimpleProduct{
			Barcode:     i,
			Description: "Must have product!",
		}

		err = db.Save(simpleProduct)
		require.NoError(t, err)
	}

	var simpleProducts []SimpleProduct

	err = db.Range("Barcode", 5, 8, &simpleProducts)
	require.NoError(t, err)

	setSkip := func(opt *index.Options) {
		opt.Skip = 2
	}
	err = db.Range("Barcode", 5, 8, &simpleProducts, setSkip)
	require.NoError(t, err)

	setZeroedLimit := func(opt *index.Options) {
		opt.Limit = 0
	}

	err = db.Range("Barcode", 5, 8, &simpleProducts, setZeroedLimit)
	require.Error(t, err)
	require.True(t, rainstorm.ErrNotFound == err)

	setLimit := func(opt *index.Options) {
		opt.Limit = 1
	}

	err = db.Range("Barcode", 5, 8, &simpleProducts, setLimit)
	require.NoError(t, err)

}

func TestIDIndexParams(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "rainstorm")
	defer os.RemoveAll(dir)
	db, _ := rainstorm.Open(filepath.Join(dir, "rainstorm.db"))
	defer db.Close()

	err := db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		require.NoError(t, err)

		idx, err := index.NewIDIndex(b, []byte("pkindex1"))
		require.NoError(t, err)

		// empty value param
		err = idx.Add([]byte(""), []byte("id"))
		require.Equal(t, index.ErrNilParam, err)

		// nil value param
		err = idx.Add(nil, []byte("id"))
		require.Equal(t, index.ErrNilParam, err)

		// empty id param
		err = idx.Add([]byte("value"), []byte(""))
		require.Equal(t, index.ErrNilParam, err)

		// nil id param
		err = idx.Add([]byte("value"), nil)
		require.Equal(t, index.ErrNilParam, err)

		// passing value and id params
		err = idx.Add([]byte("value"), []byte("id"))
		require.NoError(t, err)

		return nil
	})

	require.NoError(t, err)
}

/*func TestIDIndex(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "rainstorm")
	defer os.RemoveAll(dir)
	db, _ := rainstorm.Open(filepath.Join(dir, "rainstorm.db"))
	defer db.Close()

	err := db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		require.NoError(t, err)

		idx, err := index.NewIDIndex(b, []byte("pkindex1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hello"), []byte("id2"))
		require.Error(t, err)
		require.Equal(t, index.ErrAlreadyExists, err)

		err = idx.Add(nil, []byte("id2"))
		require.Error(t, err)
		require.Equal(t, index.ErrNilParam, err)

		err = idx.Add([]byte("hi"), nil)
		require.Error(t, err)
		require.Equal(t, index.ErrNilParam, err)

		id := idx.Get([]byte("hello"))
		require.Equal(t, []byte("id1"), id)

		id = idx.Get([]byte("goodbye"))
		require.Nil(t, id)

		err = idx.Remove([]byte("hello"))
		require.NoError(t, err)

		err = idx.Remove(nil)
		require.NoError(t, err)

		id = idx.Get([]byte("hello"))
		require.Nil(t, id)

		err = idx.Add([]byte("hello"), []byte("id1"))
		require.NoError(t, err)

		err = idx.Add([]byte("hi"), []byte("id2"))
		require.NoError(t, err)

		err = idx.Add([]byte("yo"), []byte("id3"))
		require.NoError(t, err)

		list, err := idx.AllRecords(nil)
		require.NoError(t, err)
		require.Len(t, list, 3)

		opts := index.NewOptions()
		opts.Limit = 2
		list, err = idx.AllRecords(opts)
		require.NoError(t, err)
		require.Len(t, list, 2)

		opts = index.NewOptions()
		opts.Skip = 2
		list, err = idx.AllRecords(opts)
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, []byte("id3"), list[0])

		opts = index.NewOptions()
		opts.Skip = 2
		opts.Limit = 1
		opts.Reverse = true
		list, err = idx.AllRecords(opts)
		require.NoError(t, err)
		require.Len(t, list, 1)
		require.Equal(t, []byte("id1"), list[0])

		err = idx.RemoveID([]byte("id2"))
		require.NoError(t, err)

		id = idx.Get([]byte("hello"))
		require.Equal(t, []byte("id1"), id)
		id = idx.Get([]byte("hi"))
		require.Nil(t, id)
		id = idx.Get([]byte("yo"))
		require.Equal(t, []byte("id3"), id)
		ids, err := idx.All([]byte("yo"), nil)
		require.NoError(t, err)
		require.Len(t, ids, 1)
		require.Equal(t, []byte("id3"), ids[0])

		err = idx.RemoveID([]byte("id2"))
		require.NoError(t, err)
		err = idx.RemoveID([]byte("id4"))
		require.NoError(t, err)
		return nil
	})

	require.NoError(t, err)
}

func TestIDIndexRange(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "rainstorm")
	defer os.RemoveAll(dir)
	db, _ := rainstorm.Open(filepath.Join(dir, "rainstorm.db"))
	defer db.Close()

	db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		require.NoError(t, err)

		idx, err := index.NewIDIndex(b, []byte("pkindex1"))
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			val, _ := gob.Codec.Marshal(i)
			err = idx.Add(val, val)
			require.NoError(t, err)
		}

		min, _ := gob.Codec.Marshal(3)
		max, _ := gob.Codec.Marshal(5)
		list, err := idx.Range(min, max, nil)
		require.Len(t, list, 3)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{3, 4, 5}, list)

		min, _ = gob.Codec.Marshal(11)
		max, _ = gob.Codec.Marshal(20)
		list, err = idx.Range(min, max, nil)
		require.Len(t, list, 0)
		require.NoError(t, err)

		min, _ = gob.Codec.Marshal(7)
		max, _ = gob.Codec.Marshal(2)
		list, err = idx.Range(min, max, nil)
		require.Len(t, list, 0)
		require.NoError(t, err)

		min, _ = gob.Codec.Marshal(-5)
		max, _ = gob.Codec.Marshal(2)
		list, err = idx.Range(min, max, nil)
		require.Len(t, list, 0)
		require.NoError(t, err)

		min, _ = gob.Codec.Marshal(3)
		max, _ = gob.Codec.Marshal(7)
		opts := index.NewOptions()
		opts.Skip = 2
		list, err = idx.Range(min, max, opts)
		require.Len(t, list, 3)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{5, 6, 7}, list)

		opts = index.NewOptions()
		opts.Limit = 2
		list, err = idx.Range(min, max, opts)
		require.Len(t, list, 2)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{3, 4}, list)

		opts = index.NewOptions()
		opts.Reverse = true
		opts.Skip = 2
		opts.Limit = 2
		list, err = idx.Range(min, max, opts)
		require.Len(t, list, 2)
		require.NoError(t, err)
		assertEncodedIntListEqual(t, []int{5, 4}, list)
		return nil
	})
}

func TestIDIndexPrefix(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "rainstorm")
	defer os.RemoveAll(dir)
	db, _ := rainstorm.Open(filepath.Join(dir, "rainstorm.db"))
	defer db.Close()

	db.Bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("test"))
		require.NoError(t, err)

		idx, err := index.NewIDIndex(b, []byte("pkindex1"))
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			val := []byte(fmt.Sprintf("a%d", i))
			err = idx.Add(val, val)
			require.NoError(t, err)
		}

		for i := 0; i < 10; i++ {
			val := []byte(fmt.Sprintf("b%d", i))
			err = idx.Add(val, val)
			require.NoError(t, err)
		}

		list, err := idx.Prefix([]byte("a"), nil)
		require.Len(t, list, 10)
		require.NoError(t, err)

		list, err = idx.Prefix([]byte("b"), nil)
		require.Len(t, list, 10)
		require.NoError(t, err)
		require.Equal(t, []byte("b0"), list[0])
		require.Equal(t, []byte("b9"), list[9])

		opts := index.NewOptions()
		opts.Reverse = true
		list, err = idx.Prefix([]byte("a"), opts)
		require.Len(t, list, 10)
		require.NoError(t, err)
		require.Equal(t, []byte("a9"), list[0])
		require.Equal(t, []byte("a0"), list[9])

		opts = index.NewOptions()
		opts.Reverse = true
		list, err = idx.Prefix([]byte("b"), opts)
		require.Len(t, list, 10)
		require.NoError(t, err)
		require.Equal(t, []byte("b9"), list[0])
		require.Equal(t, []byte("b0"), list[9])

		opts = index.NewOptions()
		opts.Skip = 9
		opts.Limit = 5
		list, err = idx.Prefix([]byte("a"), opts)
		require.Len(t, list, 1)
		require.NoError(t, err)
		require.Equal(t, []byte("a9"), list[0])

		opts = index.NewOptions()
		opts.Reverse = true
		opts.Skip = 9
		opts.Limit = 5
		list, err = idx.Prefix([]byte("a"), opts)
		require.Len(t, list, 1)
		require.NoError(t, err)
		require.Equal(t, []byte("a0"), list[0])
		return nil
	})
}

func assertEncodedIntListEqual(t *testing.T, expected []int, actual [][]byte) {
	ints := make([]int, len(actual))

	for i, e := range actual {
		err := gob.Codec.Unmarshal(e, &ints[i])
		require.NoError(t, err)
	}

	require.Equal(t, expected, ints)
}
*/
