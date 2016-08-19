package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func TestDropByString(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	n := db.From("b1", "b2", "b3")
	err := n.Save(&SimpleUser{ID: 10, Name: "John"})
	assert.NoError(t, err)

	err = db.From("b1").Drop("b2")
	assert.NoError(t, err)

	err = db.From("b1").Drop("b2")
	assert.Error(t, err)

	n.From("b4").Drop("b5")
	assert.Error(t, err)

	err = db.Drop("b1")
	assert.NoError(t, err)

	db.Bolt.Update(func(tx *bolt.Tx) error {
		assert.Nil(t, db.From().GetBucket(tx, "b1"))
		d := db.WithTransaction(tx)
		n := d.From("a1")
		err = n.Save(&SimpleUser{ID: 10, Name: "John"})
		assert.NoError(t, err)

		err = d.Drop("a1")
		assert.NoError(t, err)

		return nil
	})
}

func TestDropByStruct(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	n := db.From("b1", "b2", "b3")
	err := n.Save(&SimpleUser{ID: 10, Name: "John"})
	assert.NoError(t, err)

	err = n.Drop(&SimpleUser{})
	assert.NoError(t, err)

	db.Bolt.Update(func(tx *bolt.Tx) error {
		assert.Nil(t, n.GetBucket(tx, "SimpleUser"))
		d := db.WithTransaction(tx)
		n := d.From("a1")
		err = n.Save(&SimpleUser{ID: 10, Name: "John"})
		assert.NoError(t, err)

		err = n.Drop(&SimpleUser{})
		assert.NoError(t, err)

		assert.Nil(t, n.GetBucket(tx, "SimpleUser"))
		return nil
	})
}
