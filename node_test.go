package storm

import (
	"testing"

	"github.com/asdine/storm/codec/gob"
	"github.com/asdine/storm/codec/json"
	"github.com/coreos/bbolt"
	"github.com/stretchr/testify/require"
)

func TestNode(t *testing.T) {
	db, cleanup := createDB(t, Root("a"))
	defer cleanup()

	n1 := db.From("b", "c")
	node1, ok := n1.(*node)
	require.True(t, ok)
	require.Equal(t, db, node1.s)
	require.NotEqual(t, db.Node, n1)
	require.Equal(t, []string{"a"}, db.Node.(*node).rootBucket)
	require.Equal(t, []string{"a", "b", "c"}, node1.rootBucket)
	n2 := n1.From("d", "e")
	node2, ok := n2.(*node)
	require.True(t, ok)
	require.Equal(t, []string{"a", "b", "c", "d", "e"}, node2.rootBucket)
}

func TestNodeWithTransaction(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	var user User
	db.Bolt.Update(func(tx *bolt.Tx) error {
		dbx := db.WithTransaction(tx)
		err := dbx.Save(&User{ID: 10, Name: "John"})
		require.NoError(t, err)
		err = dbx.One("ID", 10, &user)
		require.NoError(t, err)
		require.Equal(t, "John", user.Name)
		return nil
	})

	err := db.One("ID", 10, &user)
	require.NoError(t, err)
}

func TestNodeWithCodec(t *testing.T) {
	t.Run("Inheritance", func(t *testing.T) {
		db, cleanup := createDB(t)
		defer cleanup()

		n := db.From("a").(*node)
		require.Equal(t, json.Codec, n.codec)
		n = n.From("b", "c", "d").(*node)
		require.Equal(t, json.Codec, n.codec)
		n = db.WithCodec(gob.Codec).(*node)
		n = n.From("e").(*node)
		require.Equal(t, gob.Codec, n.codec)
		o := n.From("f").WithCodec(json.Codec).(*node)
		require.Equal(t, gob.Codec, n.codec)
		require.Equal(t, json.Codec, o.codec)
	})

	t.Run("CodecCall", func(t *testing.T) {
		db, cleanup := createDB(t)
		defer cleanup()

		type User struct {
			ID   int
			Name string `storm:"index"`
		}

		requireBytesEqual := func(raw []byte, expected interface{}) {
			var u User
			err := gob.Codec.Unmarshal(raw, &u)
			require.NoError(t, err)
			require.Equal(t, expected, u)
		}

		n := db.From("a").WithCodec(gob.Codec)
		err := n.Set("gobBucket", "key", &User{ID: 10, Name: "John"})
		require.NoError(t, err)
		b, err := n.GetBytes("gobBucket", "key")
		require.NoError(t, err)
		requireBytesEqual(b, User{ID: 10, Name: "John"})

		id, err := toBytes(10, n.(*node).codec)
		require.NoError(t, err)

		err = n.Save(&User{ID: 10, Name: "John"})
		require.NoError(t, err)
		b, err = n.GetBytes("User", id)
		require.NoError(t, err)
		requireBytesEqual(b, User{ID: 10, Name: "John"})

		err = n.Update(&User{ID: 10, Name: "Jack"})
		require.NoError(t, err)
		b, err = n.GetBytes("User", id)
		require.NoError(t, err)
		requireBytesEqual(b, User{ID: 10, Name: "Jack"})

		err = n.UpdateField(&User{ID: 10}, "Name", "John")
		require.NoError(t, err)
		b, err = n.GetBytes("User", id)
		require.NoError(t, err)
		requireBytesEqual(b, User{ID: 10, Name: "John"})

		var users []User
		err = n.Find("Name", "John", &users)
		require.NoError(t, err)

		var user User
		err = n.One("Name", "John", &user)
		require.NoError(t, err)

		err = n.AllByIndex("Name", &users)
		require.NoError(t, err)

		err = n.All(&users)
		require.NoError(t, err)

		err = n.Range("Name", "J", "K", &users)
		require.NoError(t, err)

		err = n.Prefix("Name", "J", &users)
		require.NoError(t, err)

		_, err = n.Count(new(User))
		require.NoError(t, err)

		err = n.Select().Find(&users)
		require.NoError(t, err)
	})
}
