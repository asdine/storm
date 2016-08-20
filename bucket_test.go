package storm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucket(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	// Read tx
	readTx, err := db.Bolt.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, db.root.GetBucket(readTx, "none"))

	b, err := db.root.CreateBucketIfNotExists(readTx, "new")

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

	assert.Nil(t, db.root.GetBucket(writeTx, "none"))

	b, err = db.root.CreateBucketIfNotExists(writeTx, "new")

	assert.NoError(t, err)
	assert.NotNil(t, b)

	n2 := db.From("a", "b")
	b, err = n2.CreateBucketIfNotExists(writeTx, "c")

	assert.NoError(t, err)
	assert.NotNil(t, b)

	writeTx.Commit()

	// End write tx

	// Read tx
	readTx, err = db.Bolt.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, db.root.GetBucket(readTx, "new"))
	assert.Nil(t, db.root.GetBucket(readTx, "c"))
	assert.NotNil(t, n2.GetBucket(readTx, "c"))

	readTx.Rollback()
	// End read tx
}
