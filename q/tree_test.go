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
	assert.True(t, compare(10, 5.0, token.GTR))
}

func TestCmp(t *testing.T) {
	a := User{
		Age: 10,
	}

	b := User{
		Age: 100,
	}

	q := Eq("Age", 10)
	assert.True(t, q.Match(&a))
	assert.False(t, q.Match(&b))

	q = Gt("Age", 15)
	assert.False(t, q.Match(&a))
	assert.True(t, q.Match(&b))
}

func TestStrictEq(t *testing.T) {
	a := User{
		Age: 10,
	}

	type UserFloat struct {
		Age float64
	}

	b := UserFloat{
		Age: 10.0,
	}

	q := StrictEq("Age", 10)
	assert.True(t, q.Match(&a))
	assert.False(t, q.Match(&b))

	q = StrictEq("Age", 10.0)
	assert.False(t, q.Match(&a))
	assert.True(t, q.Match(&b))
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
	assert.True(t, q.Match(&a))
	assert.False(t, q.Match(&b))
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
	assert.True(t, q.Match(&a))
	assert.True(t, q.Match(&b))
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
	assert.True(t, q.Match(&a))
	assert.False(t, q.Match(&b))
}
