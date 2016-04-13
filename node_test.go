package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNode(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"), Root("a"))
	defer db.Close()

	n1 := db.From("b", "c")
	assert.Equal(t, db, n1.s)
	assert.NotEqual(t, db.root, n1)
	assert.Equal(t, db.root.rootBucket, []string{"a"})
	assert.Equal(t, []string{"b", "c"}, n1.rootBucket)
	n2 := n1.From("d", "e")
	assert.Equal(t, []string{"b", "c", "d", "e"}, n2.rootBucket)

}
