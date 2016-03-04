package storm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	for i := 0; i < 100; i++ {
		w := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%d", i+1)}
		err := db.Save(&w)
		assert.NoError(t, err)
	}

	err := db.All(nil)
	assert.Error(t, err)
	assert.EqualError(t, err, "provided target must be a pointer to a slice")

	var users []User

	err = db.All(&users)
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	var unknowns []UserWithNoID
	err = db.All(&unknowns)
	assert.Error(t, err)
	assert.EqualError(t, err, "bucket UserWithNoID not found")
}
