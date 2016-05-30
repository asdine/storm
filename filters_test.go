package storm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/asdine/storm/codec/gob"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestSelector(t *testing.T) {
	type User struct {
		A int32 `storm:"id"`
		B int64
		C float32
		D float64
		E string
	}

	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	tdb, _ := bolt.Open(filepath.Join(dir, "storm.db"), 0600, &bolt.Options{Timeout: 1 * time.Second})
	defer tdb.Close()

	db, _ := Open("", UseDB(tdb))

	tx, err := tdb.Begin(true)
	assert.NoError(t, err)
	defer tx.Commit()

	dbx := db.WithTransaction(tx)
	for i := 0; i < 10; i++ {
		dbx.Save(&User{
			A: int32(i),
			B: int64(i * 3),
			C: float32(i) * float32(4.3),
			D: float64(i) * float64(5.9),
			E: fmt.Sprintf("Name %d", i),
		})
	}

	var u User
	b := tx.Bucket([]byte("User"))

	err = selector(b, &filterOr{
		Left:  &filterEq{value: 5.9, field: "D"},
		Right: &filterIn{value: []interface{}{6, 12}, field: "B"},
	}, gob.Codec, &u)
	fmt.Println(err)
}
