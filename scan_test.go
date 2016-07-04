package storm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixScanDB(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	for i := 1; i < 3; i++ {
		n := db.From(fmt.Sprintf("%d%d", 2015, i))
		err := n.Save(&SimpleUser{ID: i, Name: "John"})
		assert.NoError(t, err)
	}

	for i := 1; i < 4; i++ {
		n := db.From(fmt.Sprintf("%d%d", 2016, i))
		err := n.Save(&SimpleUser{ID: i, Name: "John"})
		assert.NoError(t, err)
	}

	assert.Len(t, db.PrefixScan("2015"), 2)
	assert.Len(t, db.PrefixScan("20"), 5)

	buckets2016 := db.PrefixScan("2016")
	assert.Len(t, buckets2016, 3)
	count, err := buckets2016[1].Count(&SimpleUser{})

	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	assert.NoError(t, buckets2016[1].One("ID", 2, &SimpleUser{}))

}

func TestPrefixScanNode(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	node := db.From("node")

	for i := 1; i < 3; i++ {
		n := node.From(fmt.Sprintf("%d%d", 2016, i))
		err := n.Save(&SimpleUser{ID: i, Name: "John"})
		assert.NoError(t, err)
	}

	assert.Len(t, node.PrefixScan("2016"), 2)
}
