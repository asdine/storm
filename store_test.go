package storm_test

import (
	"testing"
	"time"

	"github.com/asdine/storm/v4"
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	t.Run("insert struct", func(t *testing.T) {
		db, err := storm.Open(":memory:")
		require.NoError(t, err)
		defer db.Close()
		s, err := db.CreateStore("foo")
		require.NoError(t, err)

		type foo struct {
			A          string
			B          int
			C          time.Time
			unexported string
		}

		f := foo{
			A:          "a",
			B:          -10,
			C:          time.Now(),
			unexported: "unexported",
		}

		id, err := s.Insert(&f)
		require.NoError(t, err)
		require.Equal(t, id, 1)
		id, err = s.Insert(&f)
		require.NoError(t, err)
		require.Equal(t, id, 2)
	})

	t.Run("insert map", func(t *testing.T) {
		db, err := storm.Open(":memory:")
		require.NoError(t, err)
		defer db.Close()
		s, err := db.CreateStore("foo")
		require.NoError(t, err)

		f := map[string]interface{}{
			"a": "a",
			"b": -10,
			"c": time.Now(),
		}

		id, err := s.Insert(&f)
		require.NoError(t, err)
		require.Equal(t, id, 1)
		id, err = s.Insert(&f)
		require.NoError(t, err)
		require.Equal(t, id, 2)
	})

	t.Run("insert json", func(t *testing.T) {
		db, err := storm.Open(":memory:")
		require.NoError(t, err)
		defer db.Close()
		s, err := db.CreateStore("foo")
		require.NoError(t, err)

		f := []byte(`{
			"a": "a",
			"b": -10
		}`)

		id, err := s.Insert(f)
		require.NoError(t, err)
		require.Equal(t, id, 1)
		id, err = s.Insert(f)
		require.NoError(t, err)
		require.Equal(t, id, 2)
	})
}

func TestAll(t *testing.T) {
	db, err := storm.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	t.Run("all", func(t *testing.T) {
		s, err := db.CreateStore("foo")
		require.NoError(t, err)

		type foo struct {
			A string
			B int
			C time.Time
		}

		c := time.Now().UTC()
		var want []foo
		for i := 0; i < 10; i++ {
			f := foo{
				A: "a",
				B: -10,
				C: c,
			}

			_, err := s.Insert(&f)
			require.NoError(t, err)
			want = append(want, f)
		}

		var got []foo
		err = s.All(&got)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}
