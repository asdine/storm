package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucket(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))
	defer db.Close()

	// Read tx
	readTx, err := db.Bolt.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, db.root.getBucket(readTx, "none"))

	b, err := db.root.createBucketIfNotExists(readTx, "new")

	// Cannot create buckets in a read transaction
	assert.Error(t, err)
	assert.Nil(t, b)

	// Read transactions in Bolt needs a rollback and not a commit
	readTx.Rollback()

	// End read tx

	// Write tx
	writeTx, err := db.Bolt.Begin(true)

	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, db.root.getBucket(writeTx, "none"))

	b, err = db.root.createBucketIfNotExists(writeTx, "new")

	assert.NoError(t, err)
	assert.NotNil(t, b)

	n2 := db.From("a", "b")
	b, err = n2.createBucketIfNotExists(writeTx, "c")

	assert.NoError(t, err)
	assert.NotNil(t, b)

	writeTx.Commit()

	// End write tx

	// Read tx
	readTx, err = db.Bolt.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, db.root.getBucket(readTx, "new"))
	assert.Nil(t, db.root.getBucket(readTx, "c"))
	assert.NotNil(t, n2.getBucket(readTx, "c"))

	readTx.Rollback()
	// End read tx
}
