package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/q"
	"github.com/stretchr/testify/assert"
)

type Score struct {
	ID    int
	Value int
}

func TestSelectFind(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"), AutoIncrement())
	defer db.Close()

	for i := 0; i < 20; i++ {
		err := db.Save(&Score{
			Value: i,
		})
		assert.NoError(t, err)
	}

	var scores []Score
	var scoresPtr []*Score

	err := db.Select(q.Eq("Value", 5)).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 1)
	assert.Equal(t, 5, scores[0].Value)

	err = db.Select(q.Eq("Value", 5)).Find(&scoresPtr)
	assert.NoError(t, err)
	assert.Len(t, scoresPtr, 1)
	assert.Equal(t, 5, scoresPtr[0].Value)

	err = db.Select(
		q.Or(
			q.Eq("Value", 5),
			q.Eq("Value", 6),
		),
	).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 2)
	assert.Equal(t, 5, scores[0].Value)
	assert.Equal(t, 6, scores[1].Value)

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 6)
	assert.Equal(t, 0, scores[0].Value)
	assert.Equal(t, 1, scores[1].Value)
	assert.Equal(t, 2, scores[2].Value)
	assert.Equal(t, 5, scores[3].Value)
	assert.Equal(t, 18, scores[4].Value)
	assert.Equal(t, 19, scores[5].Value)
}

func TestSelectFindSkip(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"), AutoIncrement())
	defer db.Close()

	for i := 0; i < 20; i++ {
		err := db.Save(&Score{
			Value: i,
		})
		assert.NoError(t, err)
	}

	var scores []Score

	err := db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(4).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 2)
	assert.Equal(t, 18, scores[0].Value)
	assert.Equal(t, 19, scores[1].Value)

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(-10).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 6)
	assert.Equal(t, 0, scores[0].Value)

	scores = nil
	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(1000).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 0)
}

func TestSelectFindLimit(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"), AutoIncrement())
	defer db.Close()

	for i := 0; i < 20; i++ {
		err := db.Save(&Score{
			Value: i,
		})
		assert.NoError(t, err)
	}

	var scores []Score

	err := db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Limit(2).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 2)
	assert.Equal(t, 0, scores[0].Value)
	assert.Equal(t, 1, scores[1].Value)

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Limit(-10).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 6)
	assert.Equal(t, 0, scores[0].Value)

	scores = nil
	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Limit(0).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 0)
}

func TestSelectFindLimitSkip(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"), AutoIncrement())
	defer db.Close()

	for i := 0; i < 20; i++ {
		err := db.Save(&Score{
			Value: i,
		})
		assert.NoError(t, err)
	}

	var scores []Score

	err := db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Limit(2).Skip(2).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 2)
	assert.Equal(t, 2, scores[0].Value)
	assert.Equal(t, 5, scores[1].Value)

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Limit(2).Skip(5).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 1)
	assert.Equal(t, 19, scores[0].Value)
}
