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

func TestCount(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	for i := 0; i < 100; i++ {
		w := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%d", i+1), DateOfBirth: time.Now().Add(-time.Duration(i*10) * time.Minute)}
		err := db.Save(&w)
		assert.NoError(t, err)
	}

	count, err := db.Count(&User{})
	assert.NoError(t, err)
	assert.Equal(t, 100, count)

	w := User{Name: "John", ID: 101, Slug: fmt.Sprintf("John%d", 101), DateOfBirth: time.Now().Add(-time.Duration(101*10) * time.Minute)}
	err = db.Save(&w)
	assert.NoError(t, err)

	count, err = db.Count(&User{})
	assert.NoError(t, err)
	assert.Equal(t, 101, count)

	tx, err := db.Begin(true)
	assert.NoError(t, err)

	count, err = tx.Count(User{})
	assert.Equal(t, ErrStructPtrNeeded, err)

	count, err = tx.Count(&User{})
	assert.NoError(t, err)
	assert.Equal(t, 101, count)

	w = User{Name: "John", ID: 102, Slug: fmt.Sprintf("John%d", 102), DateOfBirth: time.Now().Add(-time.Duration(101*10) * time.Minute)}
	err = tx.Save(&w)
	assert.NoError(t, err)

	count, err = tx.Count(&User{})
	assert.NoError(t, err)
	assert.Equal(t, 102, count)

	tx.Commit()
}
