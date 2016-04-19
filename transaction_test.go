package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	err := db.Rollback()
	assert.Error(t, err)

	err = db.Commit()
	assert.Error(t, err)

	tx, err := db.Begin(true)
	assert.NoError(t, err)

	assert.NotNil(t, tx.tx)

	err = tx.Init(&SimpleUser{})
	assert.NoError(t, err)

	err = tx.Save(&User{ID: 10, Name: "John"})
	assert.NoError(t, err)

	err = tx.Save(&User{ID: 20, Name: "John"})
	assert.NoError(t, err)

	err = tx.Save(&User{ID: 30, Name: "Steve"})
	assert.NoError(t, err)

	var user User
	err = tx.One("ID", 10, &user)
	assert.NoError(t, err)

	var users []User
	err = tx.AllByIndex("Name", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)

	err = tx.All(&users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)

	err = tx.Find("Name", "Steve", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 1)

	err = tx.Remove(&user)
	assert.NoError(t, err)

	err = tx.One("ID", 10, &user)
	assert.Error(t, err)

	err = tx.Set("b1", "best friend's mail", "mail@provider.com")
	assert.NoError(t, err)

	var str string
	err = tx.Get("b1", "best friend's mail", &str)
	assert.NoError(t, err)
	assert.Equal(t, "mail@provider.com", str)

	err = tx.Delete("b1", "best friend's mail")
	assert.NoError(t, err)

	err = tx.Get("b1", "best friend's mail", &str)
	assert.Error(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	assert.Nil(t, tx.tx)

	err = db.One("ID", 30, &user)
	assert.NoError(t, err)
	assert.Equal(t, 30, user.ID)
}

func TestTransactionRollback(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	tx, err := db.Begin(true)
	assert.NoError(t, err)

	err = tx.Save(&User{ID: 10, Name: "John"})
	assert.NoError(t, err)

	var user User
	err = tx.One("ID", 10, &user)
	assert.NoError(t, err)
	assert.Equal(t, 10, user.ID)

	err = tx.Rollback()
	assert.NoError(t, err)

	err = db.One("ID", 10, &user)
	assert.Error(t, err)
}

func TestTransactionNotWritable(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	err := db.Save(&User{ID: 10, Name: "John"})
	assert.NoError(t, err)

	tx, err := db.Begin(false)
	assert.NoError(t, err)

	err = tx.Save(&User{ID: 20, Name: "John"})
	assert.Error(t, err)

	var user User
	err = tx.One("ID", 10, &user)
	assert.NoError(t, err)

	err = tx.Rollback()
	assert.NoError(t, err)
}
