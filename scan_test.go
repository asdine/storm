package storm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrefixScan(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	node := db.From("node")

	doTestPrefixScan(t, node)
	doTestPrefixScan(t, db)

	nodeWithTransaction, _ := db.Begin(true)
	defer nodeWithTransaction.Commit()

	doTestPrefixScan(t, nodeWithTransaction)
}

func doTestPrefixScan(t *testing.T, node Node) {
	for i := 1; i < 3; i++ {
		n := node.From(fmt.Sprintf("%d%02d", 2015, i))
		err := n.Save(&SimpleUser{ID: i, Name: "John"})
		require.NoError(t, err)
	}

	for i := 1; i < 4; i++ {
		n := node.From(fmt.Sprintf("%d%02d", 2016, i))
		err := n.Save(&SimpleUser{ID: i, Name: "John"})
		require.NoError(t, err)
	}

	require.Len(t, node.PrefixScan("2015"), 2)
	require.Len(t, node.PrefixScan("20"), 5)

	buckets2016 := node.PrefixScan("2016")
	require.Len(t, buckets2016, 3)
	count, err := buckets2016[1].Count(&SimpleUser{})

	require.NoError(t, err)
	require.Equal(t, 1, count)

	require.NoError(t, buckets2016[1].One("ID", 2, &SimpleUser{}))
}

func TestPrefixScanWithEmptyPrefix(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	res := db.PrefixScan("")
	require.Len(t, res, 1)
}

func TestPrefixScanSkipValues(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	db.Set("a", "2015", 1)
	err := db.From("a", "2016").Save(&SimpleUser{ID: 1, Name: "John"})
	require.NoError(t, err)

	res := db.From("a").PrefixScan("20")
	require.Len(t, res, 1)
}

func TestRangeScan(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	node := db.From("node")

	doTestRangeScan(t, node)
	doTestRangeScan(t, db)

	nodeWithTransaction, _ := db.Begin(true)
	defer nodeWithTransaction.Commit()

	doTestRangeScan(t, nodeWithTransaction)
}

func doTestRangeScan(t *testing.T, node Node) {

	for y := 2012; y <= 2016; y++ {
		for m := 1; m <= 12; m++ {
			n := node.From(fmt.Sprintf("%d%02d", y, m))
			require.NoError(t, n.Save(&SimpleUser{ID: m, Name: "John"}))
		}
	}

	require.Len(t, node.RangeScan("2015", "2016"), 12)
	require.Len(t, node.RangeScan("201201", "201203"), 3)
	require.Len(t, node.RangeScan("2012", "201612"), 60)
	require.Len(t, node.RangeScan("2012", "2017"), 60)

	secondIn2015 := node.RangeScan("2015", "2016")[1]
	require.NoError(t, secondIn2015.One("ID", 2, &SimpleUser{}))
}

func TestRangeScanSkipValues(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	db.Set("a", "2015", 1)
	err := db.From("a", "2016").Save(&SimpleUser{ID: 1, Name: "John"})
	require.NoError(t, err)

	res := db.From("a").RangeScan("2015", "2018")
	require.Len(t, res, 1)
}
