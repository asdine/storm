package storm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAllByIndex(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	for i := 0; i < 100; i++ {
		w := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%d", i+1), DateOfBirth: time.Now().Add(-time.Duration(i*10) * time.Minute)}
		err := db.Save(&w)
		assert.NoError(t, err)
	}

	err := db.AllByIndex("", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrSlicePtrNeeded, err)

	var users []User

	err = db.AllByIndex("DateOfBirth", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 100, users[0].ID)
	assert.Equal(t, 1, users[99].ID)

	err = db.AllByIndex("Name", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	y := UniqueNameUser{Name: "Jake", ID: 200}
	err = db.Save(&y)
	assert.NoError(t, err)

	var y2 []UniqueNameUser
	err = db.AllByIndex("ID", &y2)
	assert.NoError(t, err)
	assert.Len(t, y2, 1)

	n := NestedID{}
	n.ID = "100"
	n.Name = "John"

	err = db.Save(&n)
	assert.NoError(t, err)

	var n2 []NestedID
	err = db.AllByIndex("ID", &n2)
	assert.NoError(t, err)
	assert.Len(t, n2, 1)

	err = db.AllByIndex("Name", &users, Limit(10))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 10, users[9].ID)

	err = db.AllByIndex("Name", &users, Limit(200))
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	err = db.AllByIndex("Name", &users, Limit(-10))
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	err = db.AllByIndex("Name", &users, Skip(200))
	assert.NoError(t, err)
	assert.Len(t, users, 0)

	err = db.AllByIndex("Name", &users, Skip(-10))
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	err = db.AllByIndex("ID", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	err = db.AllByIndex("ID", &users, Limit(10))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 10, users[9].ID)

	err = db.AllByIndex("ID", &users, Skip(10))
	assert.NoError(t, err)
	assert.Len(t, users, 90)
	assert.Equal(t, 11, users[0].ID)
	assert.Equal(t, 100, users[89].ID)

	err = db.AllByIndex("Name", &users, Limit(10), Skip(10))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 11, users[0].ID)
	assert.Equal(t, 20, users[9].ID)
}

func TestAll(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	for i := 0; i < 100; i++ {
		w := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%d", i+1), DateOfBirth: time.Now().Add(-time.Duration(i*10) * time.Minute)}
		err := db.Save(&w)
		assert.NoError(t, err)
	}

	var users []User

	err := db.All(&users)
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	var users2 []*User

	err = db.All(&users2)
	assert.NoError(t, err)
	assert.Len(t, users2, 100)
	assert.Equal(t, 1, users2[0].ID)
	assert.Equal(t, 100, users2[99].ID)

	var unknowns []UserWithNoID
	err = db.All(&unknowns)
	assert.Error(t, err)
	assert.Equal(t, ErrNoID, err)

	err = db.Save(&NestedID{
		ToEmbed: ToEmbed{ID: "id1"},
		Name:    "John",
	})
	assert.NoError(t, err)

	err = db.Save(&NestedID{
		ToEmbed: ToEmbed{ID: "id2"},
		Name:    "Mike",
	})
	assert.NoError(t, err)

	db.Save(&NestedID{
		ToEmbed: ToEmbed{ID: "id3"},
		Name:    "Steve",
	})
	assert.NoError(t, err)

	var nested []NestedID
	err = db.All(&nested)
	assert.NoError(t, err)
	assert.Len(t, nested, 3)

	err = db.All(&users, Limit(10), Skip(10))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 11, users[0].ID)
	assert.Equal(t, 20, users[9].ID)
}
