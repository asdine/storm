package storm

import (
	"fmt"
	"testing"

	"github.com/asdine/storm/codec/gob"
	"github.com/asdine/storm/q"
	"github.com/stretchr/testify/assert"
)

type Score struct {
	ID    int
	Value int
}

func prepareScoreDB(t *testing.T) (*DB, func()) {
	db, cleanup := createDB(t, AutoIncrement(), Codec(gob.Codec))

	for i := 0; i < 20; i++ {
		err := db.Save(&Score{
			Value: i,
		})
		assert.NoError(t, err)
	}

	return db, cleanup
}

func TestSelectFind(t *testing.T) {
	db, cleanup := prepareScoreDB(t)
	defer cleanup()

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

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Reverse().Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 6)
	assert.Equal(t, 19, scores[0].Value)
	assert.Equal(t, 18, scores[1].Value)
	assert.Equal(t, 5, scores[2].Value)
	assert.Equal(t, 2, scores[3].Value)
	assert.Equal(t, 1, scores[4].Value)
	assert.Equal(t, 0, scores[5].Value)
}

func TestSelectFindSkip(t *testing.T) {
	db, cleanup := prepareScoreDB(t)
	defer cleanup()

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
	assert.Error(t, err)
	assert.True(t, ErrNotFound == err)
	assert.Len(t, scores, 0)
}

func TestSelectFindLimit(t *testing.T) {
	db, cleanup := prepareScoreDB(t)
	defer cleanup()
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
	assert.Error(t, err)
	assert.True(t, ErrNotFound == err)
	assert.Len(t, scores, 0)
}

func TestSelectFindLimitSkip(t *testing.T) {
	db, cleanup := prepareScoreDB(t)
	defer cleanup()

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

func TestSelectFirst(t *testing.T) {
	db, cleanup := prepareScoreDB(t)
	defer cleanup()

	var score Score

	err := db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(2).First(&score)
	assert.NoError(t, err)
	assert.Equal(t, 2, score.Value)

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(1).Reverse().First(&score)
	assert.NoError(t, err)
	assert.Equal(t, 18, score.Value)
}

func TestSelectRemove(t *testing.T) {
	db, cleanup := prepareScoreDB(t)
	defer cleanup()

	err := db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(2).Delete(&Score{})
	assert.NoError(t, err)

	var scores []Score
	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Find(&scores)
	assert.NoError(t, err)
	assert.Len(t, scores, 2)
	assert.Equal(t, 0, scores[0].Value)
	assert.Equal(t, 1, scores[1].Value)

	for i := 0; i < 10; i++ {
		w := User{ID: i + 1, Name: fmt.Sprintf("John%d", i+1)}
		err = db.Save(&w)
		assert.NoError(t, err)
	}

	err = db.Select(q.Gte("ID", 5)).Delete(&User{})
	assert.NoError(t, err)

	var user User
	err = db.One("Name", "John6", &user)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	err = db.One("Name", "John4", &user)
	assert.NoError(t, err)
}

func TestSelectCount(t *testing.T) {
	db, cleanup := prepareScoreDB(t)
	defer cleanup()

	total, err := db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Count(&Score{})
	assert.NoError(t, err)
	assert.Equal(t, 6, total)

	total, err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(2).Count(&Score{})
	assert.NoError(t, err)
	assert.Equal(t, 4, total)

	total, err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(2).Limit(2).Count(&Score{})
	assert.NoError(t, err)
	assert.Equal(t, 2, total)

	total, err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(5).Limit(2).Count(&Score{})
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
}

// func TestSelectRaw(t *testing.T) {
// 	db, cleanup := createDB(t, AutoIncrement(), Codec(json.Codec))
// 	defer cleanup()
//
// 	for i := 0; i < 20; i++ {
// 		err := db.Save(&Score{
// 			Value: i,
// 		})
// 		assert.NoError(t, err)
// 	}
//
// 	list, err := db.Select(q.Gte("Value", 18)).Raw("Score")
// 	assert.NoError(t, err)
// 	assert.Len(t, list, 2)
// }
