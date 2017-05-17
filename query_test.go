package storm

import (
	"fmt"
	"testing"

	"github.com/asdine/storm/codec/json"
	"github.com/asdine/storm/q"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Score struct {
	ID    int
	Value int
}

func prepareScoreDB(t *testing.T) (*DB, func()) {
	db, cleanup := createDB(t, AutoIncrement())

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

func TestSelectFindOrderBy(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	type T struct {
		ID  int `storm:"increment"`
		Str string
		Int int
		Rnd int
	}

	strs := []string{"e", "b", "d", "a", "c", "d"}
	ints := []int{2, 3, 5, 4, 2, 1}
	for i := 0; i < 6; i++ {
		record := T{
			Str: strs[i],
			Int: ints[i],
		}
		if i == 3 {
			record.Rnd = 3
		}

		err := db.Save(&record)
		assert.NoError(t, err)
	}

	var list []T
	err := db.Select().OrderBy("ID").Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 6)
	for i, j := 0, 0; i < 6; i, j = i+1, j+1 {
		if i == 2 {
			j--
		}
		assert.Equal(t, i+1, list[i].ID)
	}

	err = db.Select().OrderBy("Str").Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 6)
	for i, j := 0, 0; i < 6; i, j = i+1, j+1 {
		if i == 4 {
			j--
		}
		assert.Equal(t, string([]byte{'a' + byte(j)}), list[i].Str)
	}

	err = db.Select().OrderBy("Int").Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 6)
	for i, j := 0, 0; i < 6; i, j = i+1, j+1 {
		if i == 2 {
			j--
		}
		assert.Equal(t, j+1, list[i].Int)
	}

	err = db.Select().OrderBy("Rnd").Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 6)
	assert.Equal(t, 1, list[0].ID)
	assert.Equal(t, 2, list[1].ID)
	assert.Equal(t, 3, list[2].ID)
	assert.Equal(t, 5, list[3].ID)
	assert.Equal(t, 6, list[4].ID)
	assert.Equal(t, 4, list[5].ID)

	err = db.Select().OrderBy("Int").Reverse().Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 6)
	for i, j := 0, 0; i < 6; i, j = i+1, j+1 {
		if i == 4 {
			j--
		}
		assert.Equal(t, 5-j, list[i].Int)
	}

	err = db.Select().OrderBy("Int").Reverse().Limit(2).Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	for i := 0; i < 2; i++ {
		assert.Equal(t, 5-i, list[i].Int)
	}

	err = db.Select().OrderBy("Int").Reverse().Skip(2).Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 4)
	for i, j := 0, 0; i < 3; i, j = i+1, j+1 {
		if i == 2 {
			j--
		}
		assert.Equal(t, 3-j, list[i].Int)
	}

	err = db.Select().OrderBy("Int").Reverse().Skip(5).Limit(2).Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, 1, list[0].Int)

	err = db.Select().OrderBy("Str", "Int").Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 6)
	assert.Equal(t, "a", list[0].Str)
	assert.Equal(t, 4, list[0].Int)
	assert.Equal(t, "b", list[1].Str)
	assert.Equal(t, 3, list[1].Int)
	assert.Equal(t, "c", list[2].Str)
	assert.Equal(t, 2, list[2].Int)
	assert.Equal(t, "d", list[3].Str)
	assert.Equal(t, 1, list[3].Int)
	assert.Equal(t, "d", list[4].Str)
	assert.Equal(t, 5, list[4].Int)
	assert.Equal(t, "e", list[5].Str)
	assert.Equal(t, 2, list[5].Int)
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

func TestSelectFirstOrderBy(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	type T struct {
		ID  int `storm:"increment"`
		Str string
		Int int
	}

	strs := []string{"e", "b", "a", "c", "d"}
	ints := []int{2, 3, 1, 4, 5}
	for i := 0; i < 5; i++ {
		err := db.Save(&T{
			Str: strs[i],
			Int: ints[i],
		})
		assert.NoError(t, err)
	}

	var record T
	err := db.Select().OrderBy("ID").First(&record)
	assert.NoError(t, err)
	assert.Equal(t, 1, record.ID)

	err = db.Select().OrderBy("Str").First(&record)
	assert.NoError(t, err)
	assert.Equal(t, "a", record.Str)

	err = db.Select().OrderBy("Int").First(&record)
	assert.NoError(t, err)
	assert.Equal(t, 1, record.Int)

	err = db.Select().OrderBy("Int").Reverse().First(&record)
	assert.NoError(t, err)
	assert.Equal(t, 5, record.Int)

	err = db.Select().OrderBy("Int").Reverse().Limit(2).First(&record)
	assert.NoError(t, err)
	assert.Equal(t, 5, record.Int)

	err = db.Select().OrderBy("Int").Reverse().Skip(2).First(&record)
	assert.NoError(t, err)
	assert.Equal(t, 3, record.Int)

	err = db.Select().OrderBy("Int").Reverse().Skip(4).Limit(2).First(&record)
	assert.NoError(t, err)
	assert.Equal(t, 1, record.Int)
}

func TestSelectDelete(t *testing.T) {
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

func TestSelectRaw(t *testing.T) {
	db, cleanup := createDB(t, AutoIncrement(), Codec(json.Codec))
	defer cleanup()

	for i := 0; i < 20; i++ {
		err := db.Save(&Score{
			Value: i,
		})
		assert.NoError(t, err)
	}

	list, err := db.Select().Bucket("Score").Raw()
	assert.NoError(t, err)
	assert.Len(t, list, 20)

	list, err = db.Select().Bucket("Score").Skip(18).Limit(5).Raw()
	assert.NoError(t, err)
	assert.Len(t, list, 2)

	i := 0
	err = db.Select().Bucket("Score").Skip(18).Limit(5).RawEach(func(k []byte, v []byte) error {
		i++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, i, 2)
}

func TestSelectEach(t *testing.T) {
	db, cleanup := createDB(t, AutoIncrement(), Codec(json.Codec))
	defer cleanup()

	for i := 0; i < 20; i++ {
		err := db.Save(&Score{
			Value: i,
		})
		assert.NoError(t, err)
	}

	i := 0
	err := db.Select().Each(new(Score), func(record interface{}) error {
		s, ok := record.(*Score)
		require.True(t, ok)
		require.Equal(t, i, s.Value)
		i++
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 20, i)

	i = 0
	err = db.Select().Skip(18).Limit(5).Each(new(Score), func(record interface{}) error {
		s, ok := record.(*Score)
		require.True(t, ok)
		require.Equal(t, i+18, s.Value)
		i++
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 2, i)
}
