package storm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	for i := 0; i < 100; i++ {
		w := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%d", i+1)}

		if i%2 == 0 {
			w.Group = "staff"
		} else {
			w.Group = "normal"
		}

		err := db.Save(&w)
		assert.NoError(t, err)
	}

	err := db.Find("Name", "John", &User{})
	assert.Error(t, err)
	assert.Equal(t, ErrSlicePtrNeeded, err)

	err = db.Find("Name", "John", &[]struct {
		Name string
		ID   int
	}{})
	assert.Error(t, err)
	assert.Equal(t, ErrNoName, err)

	notTheRightUsers := []UniqueNameUser{}

	err = db.Find("Name", "John", &notTheRightUsers)
	assert.Error(t, err)
	assert.EqualError(t, err, "not found")

	users := []User{}

	err = db.Find("Age", "John", &users)
	assert.Error(t, err)
	assert.EqualError(t, err, "field Age not found")

	err = db.Find("DateOfBirth", "John", &users)
	assert.Error(t, err)
	assert.True(t, ErrNotFound == err)

	err = db.Find("Group", "staff", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 50)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 99, users[49].ID)

	err = db.Find("Group", "admin", &users)
	assert.Error(t, err)
	assert.True(t, ErrNotFound == err)

	err = db.Find("Name", "John", users)
	assert.Error(t, err)
	assert.Equal(t, ErrSlicePtrNeeded, err)

	err = db.Find("Name", "John", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	users = []User{}
	err = db.Find("Slug", "John10", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, 10, users[0].ID)

	users = []User{}
	err = db.Find("Name", nil, &users)
	assert.Error(t, err)
	assert.True(t, ErrNotFound == err)

	err = db.Find("Name", "John", &users, Limit(10), Skip(20))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 21, users[0].ID)
	assert.Equal(t, 30, users[9].ID)
}
