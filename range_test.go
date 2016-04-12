package storm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRange(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	for i := 0; i < 100; i++ {
		w := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%03d", i+1)}
		err := db.Save(&w)
		assert.NoError(t, err)
		z := User{Name: fmt.Sprintf("Zach%03d", i+1), ID: i + 101, Slug: fmt.Sprintf("Zach%03d", i+1)}
		err = db.Save(&z)
		assert.NoError(t, err)
	}

	min := "John010"
	max := "John020"
	var users []User
	err := db.Range("Slug", min, max, &users)
	assert.NoError(t, err)
	assert.Len(t, users, 11)
	assert.Equal(t, "John010", users[0].Slug)
	assert.Equal(t, "John020", users[10].Slug)

	min = "Zach010"
	max = "Zach020"
	users = nil
	err = db.Range("Name", min, max, &users)
	assert.NoError(t, err)
	assert.Len(t, users, 11)
	assert.Equal(t, "Zach010", users[0].Name)
	assert.Equal(t, "Zach020", users[10].Name)

	err = db.Range("Name", min, max, &User{})
	assert.Error(t, err)
	assert.Equal(t, ErrSlicePtrNeeded, err)

	notTheRightUsers := []UniqueNameUser{}

	err = db.Range("Name", min, max, &notTheRightUsers)
	assert.Error(t, err)
	assert.EqualError(t, err, "bucket UniqueNameUser not found")

	users = nil

	err = db.Range("Age", min, max, &users)
	assert.Error(t, err)
	assert.EqualError(t, err, "field Age not found")

	err = db.Range("DateOfBirth", min, max, &users)
	assert.NoError(t, err)

	err = db.Range("Group", min, max, &users)
	assert.Error(t, err)
	assert.EqualError(t, err, "index Group not found")
}
