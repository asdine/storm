package storm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixScan(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	node := db.From("node")

	doTestPrefixScan(t, node)
	doTestPrefixScan(t, db)

	nodeWithTransaction, _ := db.Begin(true)
	defer nodeWithTransaction.Commit()

	doTestPrefixScan(t, nodeWithTransaction)
}

func doTestPrefixScan(t *testing.T, node bucketScanner) {
	for i := 1; i < 3; i++ {
		n := node.From(fmt.Sprintf("%d%02d", 2015, i))
		err := n.Save(&SimpleUser{ID: i, Name: "John"})
		assert.NoError(t, err)
	}

	for i := 1; i < 4; i++ {
		n := node.From(fmt.Sprintf("%d%02d", 2016, i))
		err := n.Save(&SimpleUser{ID: i, Name: "John"})
		assert.NoError(t, err)
	}

	assert.Len(t, node.PrefixScan("2015"), 2)
	assert.Len(t, node.PrefixScan("20"), 5)

	buckets2016 := node.PrefixScan("2016")
	assert.Len(t, buckets2016, 3)
	count, err := buckets2016[1].Count(&SimpleUser{})

	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	assert.NoError(t, buckets2016[1].One("ID", 2, &SimpleUser{}))
}

func TestRangeScan(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	node := db.From("node")

	doTestRangeScan(t, node)
	doTestRangeScan(t, db)

	nodeWithTransaction, _ := db.Begin(true)
	defer nodeWithTransaction.Commit()

	doTestRangeScan(t, nodeWithTransaction)
}

func doTestRangeScan(t *testing.T, node bucketScanner) {

	for y := 2012; y <= 2016; y++ {
		for m := 1; m <= 12; m++ {
			n := node.From(fmt.Sprintf("%d%02d", y, m))
			assert.NoError(t, n.Save(&SimpleUser{ID: m, Name: "John"}))
		}
	}

	assert.Len(t, node.RangeScan("2015", "2016"), 12)
	assert.Len(t, node.RangeScan("201201", "201203"), 3)
	assert.Len(t, node.RangeScan("2012", "201612"), 60)
	assert.Len(t, node.RangeScan("2012", "2017"), 60)

	secondIn2015 := node.RangeScan("2015", "2016")[1]
	assert.NoError(t, secondIn2015.One("ID", 2, &SimpleUser{}))
}

type bucketScanner interface {
	From(addend ...string) *Node
	PrefixScan(prefix string) []*Node
	RangeScan(min, max string) []*Node
}
