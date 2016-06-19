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

func TestRemove(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	u1 := IndexedNameUser{ID: 10, Name: "John", age: 10}
	err := db.Save(&u1)
	assert.NoError(t, err)

	err = db.Remove(u1)
	assert.Equal(t, ErrStructPtrNeeded, err)

	err = db.Remove(&u1)
	assert.NoError(t, err)

	err = db.Remove(&u1)
	assert.Equal(t, ErrNotFound, err)

	u2 := IndexedNameUser{}
	err = db.Get("IndexedNameUser", 10, &u2)
	assert.True(t, ErrNotFound == err)

	err = db.Remove(nil)
	assert.Equal(t, ErrStructPtrNeeded, err)

	var users []User
	for i := 0; i < 10; i++ {
		user := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%d", i+1), DateOfBirth: time.Now().Add(-time.Duration(i*10) * time.Minute)}
		err := db.Save(&user)
		assert.NoError(t, err)
		users = append(users, user)
	}

	err = db.Remove(&users[0])
	assert.NoError(t, err)
	err = db.Remove(&users[1])
	assert.NoError(t, err)

	users = nil
	err = db.All(&users)
	assert.NoError(t, err)
	assert.Len(t, users, 8)
	assert.Equal(t, 3, users[0].ID)
}
