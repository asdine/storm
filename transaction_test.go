package storm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransaction(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	err := db.Rollback()
	require.Error(t, err)

	err = db.Commit()
	require.Error(t, err)

	tx, err := db.Begin(true)
	require.NoError(t, err)

	ntx, ok := tx.(*node)
	require.True(t, ok)
	require.NotNil(t, ntx.tx)

	err = tx.Init(&SimpleUser{})
	require.NoError(t, err)

	err = tx.Save(&User{ID: 10, Name: "John"})
	require.NoError(t, err)

	err = tx.Save(&User{ID: 20, Name: "John"})
	require.NoError(t, err)

	err = tx.Save(&User{ID: 30, Name: "Steve"})
	require.NoError(t, err)

	var user User
	err = tx.One("ID", 10, &user)
	require.NoError(t, err)

	var users []User
	err = tx.AllByIndex("Name", &users)
	require.NoError(t, err)
	require.Len(t, users, 3)

	err = tx.All(&users)
	require.NoError(t, err)
	require.Len(t, users, 3)

	err = tx.Find("Name", "Steve", &users)
	require.NoError(t, err)
	require.Len(t, users, 1)

	err = tx.DeleteStruct(&user)
	require.NoError(t, err)

	err = tx.One("ID", 10, &user)
	require.Error(t, err)

	err = tx.Set("b1", "best friend's mail", "mail@provider.com")
	require.NoError(t, err)

	var str string
	err = tx.Get("b1", "best friend's mail", &str)
	require.NoError(t, err)
	require.Equal(t, "mail@provider.com", str)

	err = tx.Delete("b1", "best friend's mail")
	require.NoError(t, err)

	err = tx.Get("b1", "best friend's mail", &str)
	require.Error(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	err = tx.Commit()
	require.Error(t, err)
	require.Equal(t, ErrNotInTransaction, err)

	err = db.One("ID", 30, &user)
	require.NoError(t, err)
	require.Equal(t, 30, user.ID)
}

func TestTransactionRollback(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	tx, err := db.Begin(true)
	require.NoError(t, err)

	err = tx.Save(&User{ID: 10, Name: "John"})
	require.NoError(t, err)

	var user User
	err = tx.One("ID", 10, &user)
	require.NoError(t, err)
	require.Equal(t, 10, user.ID)

	err = tx.Rollback()
	require.NoError(t, err)

	err = db.One("ID", 10, &user)
	require.Error(t, err)
}

func TestTransactionNotWritable(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	err := db.Save(&User{ID: 10, Name: "John"})
	require.NoError(t, err)

	tx, err := db.Begin(false)
	require.NoError(t, err)

	err = tx.Save(&User{ID: 20, Name: "John"})
	require.Error(t, err)

	var user User
	err = tx.One("ID", 10, &user)
	require.NoError(t, err)

	err = tx.Rollback()
	require.NoError(t, err)
}
