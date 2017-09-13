package storm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBucket(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	// Read tx
	readTx, err := db.Bolt.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	require.Nil(t, db.GetBucket(readTx, "none"))

	b, err := db.CreateBucketIfNotExists(readTx, "new")

	// Cannot create buckets in a read transaction
	require.Error(t, err)
	require.Nil(t, b)

	// Read transactions in Bolt needs a rollback and not a commit
	readTx.Rollback()

	// End read tx

	// Write tx
	writeTx, err := db.Bolt.Begin(true)

	if err != nil {
		t.Fatal(err)
	}

	require.Nil(t, db.GetBucket(writeTx, "none"))

	b, err = db.CreateBucketIfNotExists(writeTx, "new")

	require.NoError(t, err)
	require.NotNil(t, b)

	n2 := db.From("a", "b")
	b, err = n2.CreateBucketIfNotExists(writeTx, "c")

	require.NoError(t, err)
	require.NotNil(t, b)

	writeTx.Commit()

	// End write tx

	// Read tx
	readTx, err = db.Bolt.Begin(false)
	if err != nil {
		t.Fatal(err)
	}

	require.NotNil(t, db.GetBucket(readTx, "new"))
	require.Nil(t, db.GetBucket(readTx, "c"))
	require.NotNil(t, n2.GetBucket(readTx, "c"))

	readTx.Rollback()
	// End read tx
}
