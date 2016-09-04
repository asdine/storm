package storm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	err := db.Rollback()
	assert.Error(t, err)

	err = db.Commit()
	assert.Error(t, err)

	tx, err := db.Begin(true)
	assert.NoError(t, err)

	ntx, ok := tx.(*node)
	assert.True(t, ok)
	assert.NotNil(t, ntx.tx)

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

	err = tx.DeleteStruct(&user)
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

	err = tx.Commit()
	assert.Error(t, err)
	assert.Equal(t, ErrNotInTransaction, err)

	err = db.One("ID", 30, &user)
	assert.NoError(t, err)
	assert.Equal(t, 30, user.ID)
}

func TestTransactionRollback(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

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
	db, cleanup := createDB(t)
	defer cleanup()

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
