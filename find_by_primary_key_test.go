package storm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindByPrimaryKey(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	for i := 0; i < 100; i++ {
		w := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%d", i+1)}
		err := db.Save(&w)
		assert.NoError(t, err)
	}

	user := User{}
	err := db.FindByPrimaryKey(10, &user)
	assert.NoError(t, err)
	assert.Equal(t, user.Slug, "John10")
}
