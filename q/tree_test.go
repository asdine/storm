package q

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	Age  int
	Name string
}

func TestCompare(t *testing.T) {
	assert.True(t, compare(10, 10, token.EQL))
	assert.True(t, compare(10, 10.0, token.EQL))
	assert.True(t, compare(10, "10", token.EQL))
	assert.True(t, compare(10, "10.0", token.EQL))
	assert.False(t, compare(10, "hello", token.EQL))
	assert.True(t, compare(10.0, 10, token.EQL))
	assert.True(t, compare(10.0, 10.0, token.EQL))
	assert.True(t, compare(10.0, "10", token.EQL))
	assert.True(t, compare(10.0, "10.0", token.EQL))
	assert.False(t, compare(10.0, "hello", token.EQL))
	assert.True(t, compare("hello", "hello", token.EQL))
	assert.True(t, compare(&User{Name: "John"}, &User{Name: "John"}, token.EQL))
	assert.False(t, compare(&User{Name: "John"}, &User{Name: "Jack"}, token.GTR))
}

func TestEq(t *testing.T) {
	a := User{
		Age: 10,
	}

	b := User{
		Age: 100,
	}

	q := Eq("Age", 10)
	assert.True(t, q.Exec(&a))
	assert.False(t, q.Exec(&b))
}

func TestAnd(t *testing.T) {
	a := User{
		Age:  10,
		Name: "John",
	}

	b := User{
		Age:  10,
		Name: "Jack",
	}

	q := And(
		Eq("Age", 10),
		Eq("Name", "John"),
	)
	assert.True(t, q.Exec(&a))
	assert.False(t, q.Exec(&b))
}

func TestOr(t *testing.T) {
	a := User{
		Age:  10,
		Name: "John",
	}

	b := User{
		Age:  10,
		Name: "Jack",
	}

	q := Or(
		Eq("Age", 10),
		Eq("Name", "Jack"),
	)
	assert.True(t, q.Exec(&a))
	assert.True(t, q.Exec(&b))
}

func TestAndOr(t *testing.T) {
	a := User{
		Age:  10,
		Name: "John",
	}

	b := User{
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
	assert.True(t, q.Exec(&a))
	assert.False(t, q.Exec(&b))
}
