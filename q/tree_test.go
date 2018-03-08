package q

import (
	"go/token"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type A struct {
	Age  int
	Name string
}

type B struct {
	Age  int
	Name string
}

func TestCompare(t *testing.T) {
	require.True(t, compare(10, 10, token.EQL))
	require.True(t, compare(10, 10.0, token.EQL))
	require.True(t, compare(10, "10", token.EQL))
	require.True(t, compare(10, "10.0", token.EQL))
	require.False(t, compare(10, "hello", token.EQL))
	require.True(t, compare(10.0, 10, token.EQL))
	require.True(t, compare(10.0, 10.0, token.EQL))
	require.True(t, compare(10.0, "10", token.EQL))
	require.True(t, compare(10.0, "10.0", token.EQL))
	require.False(t, compare(10.0, "hello", token.EQL))
	require.True(t, compare("hello", "hello", token.EQL))
	require.True(t, compare(&A{Name: "John"}, &A{Name: "John"}, token.EQL))
	require.False(t, compare(&A{Name: "John"}, &A{Name: "Jack"}, token.GTR))
	require.True(t, compare(10, 5.0, token.GTR))
	t1 := time.Now()
	t2 := t1.Add(2 * time.Hour)
	t3 := t1.Add(-2 * time.Hour)
	require.True(t, compare(t1, t1, token.EQL))
	require.True(t, compare(t1, t2, token.LSS))
	require.True(t, compare(t1, t3, token.GTR))
	require.False(t, compare(&A{Name: "John"}, t1, token.EQL))
	require.False(t, compare(&A{Name: "John"}, t1, token.LEQ))
	require.True(t, compare(uint32(10), uint32(5), token.GTR))
	require.False(t, compare(uint32(5), uint32(10), token.GTR))
	require.True(t, compare(uint32(10), int32(5), token.GTR))
	require.True(t, compare(uint32(10), float32(5), token.GTR))
	require.True(t, compare(int32(10), uint32(5), token.GTR))
	require.True(t, compare(float32(10), uint32(5), token.GTR))
}

func TestCmp(t *testing.T) {
	a := A{
		Age: 10,
	}

	b := A{
		Age: 100,
	}

	q := Eq("Age", 10)
	ok, err := q.Match(&a)
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = q.Match(&b)
	require.NoError(t, err)
	require.False(t, ok)

	q = Gt("Age", 15)
	ok, err = q.Match(&a)
	require.NoError(t, err)
	require.False(t, ok)
	ok, err = q.Match(&b)
	require.NoError(t, err)
	require.True(t, ok)

	// Unknown field
	q = Gt("Unknown", 15)
	ok, err = q.Match(&a)
	require.Equal(t, err, ErrUnknownField)
	require.False(t, ok)

	q = EqF("Age", "Age")
	ok, err = q.Match(&a)
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = q.Match(&b)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestStrictEq(t *testing.T) {
	a := A{
		Age: 10,
	}

	type UserFloat struct {
		Age float64
	}

	b := UserFloat{
		Age: 10.0,
	}

	q := StrictEq("Age", 10)
	ok, err := q.Match(&a)
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = q.Match(&b)
	require.NoError(t, err)
	require.False(t, ok)

	q = StrictEq("Age", 10.0)
	ok, err = q.Match(&a)
	require.NoError(t, err)
	require.False(t, ok)
	ok, err = q.Match(&b)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestIn(t *testing.T) {
	a := A{
		Age: 10,
	}

	q := In("Age", []int{1, 5, 10, 3})
	ok, err := q.Match(&a)
	require.NoError(t, err)
	require.True(t, ok)

	q = In("Age", []int{1, 5, 3})
	ok, err = q.Match(&a)
	require.NoError(t, err)
	require.False(t, ok)

	q = In("Age", []int{})
	ok, err = q.Match(&a)
	require.NoError(t, err)
	require.False(t, ok)

	q = In("Age", nil)
	ok, err = q.Match(&a)
	require.NoError(t, err)
	require.False(t, ok)

	q = In("Age", []float64{1.0, 5.0, 10.0, 3.0})
	ok, err = q.Match(&a)
	require.NoError(t, err)
	require.True(t, ok)

	q = In("Age", 10)
	ok, err = q.Match(&a)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestAnd(t *testing.T) {
	a := A{
		Age:  10,
		Name: "John",
	}

	b := A{
		Age:  10,
		Name: "Jack",
	}

	q := And(
		Eq("Age", 10),
		Eq("Name", "John"),
	)
	ok, err := q.Match(&a)
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = q.Match(&b)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestOr(t *testing.T) {
	a := A{
		Age:  10,
		Name: "John",
	}

	b := A{
		Age:  10,
		Name: "Jack",
	}

	q := Or(
		Eq("Age", 10),
		Eq("Name", "Jack"),
	)
	ok, err := q.Match(&a)
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = q.Match(&b)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestNot(t *testing.T) {
	q := Not(
		Eq("Age", 10),
	)
	ok, err := q.Match(&A{
		Age: 11,
	})
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = q.Match(&A{
		Age: 10,
	})
	require.NoError(t, err)
	require.False(t, ok)

	q = Not(
		Gt("Age", 10),
		Eq("Name", "John"),
	)
	ok, err = q.Match(&A{
		Age: 8,
	})
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = q.Match(&A{
		Age:  11,
		Name: "Jack",
	})
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = q.Match(&A{
		Age:  5,
		Name: "John",
	})
	require.NoError(t, err)
	require.False(t, ok)
}

func TestAndOr(t *testing.T) {
	a := A{
		Age:  10,
		Name: "John",
	}

	b := A{
		Age:  100,
		Name: "Jack",
	}

	q := And(
		Eq("Age", 10),
		Or(
			Eq("Name", "Jack"),
			Eq("Name", "John"),
		),
	)
	ok, err := q.Match(&a)
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = q.Match(&b)
	require.NoError(t, err)
	require.False(t, ok)
}
