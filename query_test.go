package storm

import (
	"fmt"
	"testing"

	"github.com/asdine/storm/codec/json"
	"github.com/asdine/storm/q"
	"github.com/stretchr/testify/require"
)

type Score struct {
	ID    int `storm:"increment"`
	Value int
}

func prepareScoreDB(t *testing.T) (*DB, func()) {
	db, cleanup := createDB(t)

	for i := 0; i < 20; i++ {
		err := db.Save(&Score{
			Value: i,
		})
		require.NoError(t, err)
	}

	return db, cleanup
}

func TestSelectFind(t *testing.T) {
	db, cleanup := prepareScoreDB(t)
	defer cleanup()

	var scores []Score
	var scoresPtr []*Score

	err := db.Select(q.Eq("Value", 5)).Find(&scores)
	require.NoError(t, err)
	require.Len(t, scores, 1)
	require.Equal(t, 5, scores[0].Value)

	err = db.Select(q.Eq("Value", 5)).Find(&scoresPtr)
	require.NoError(t, err)
	require.Len(t, scoresPtr, 1)
	require.Equal(t, 5, scoresPtr[0].Value)

	err = db.Select(
		q.Or(
			q.Eq("Value", 5),
			q.Eq("Value", 6),
		),
	).Find(&scores)
	require.NoError(t, err)
	require.Len(t, scores, 2)
	require.Equal(t, 5, scores[0].Value)
	require.Equal(t, 6, scores[1].Value)

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Find(&scores)
	require.NoError(t, err)
	require.Len(t, scores, 6)
	require.Equal(t, 0, scores[0].Value)
	require.Equal(t, 1, scores[1].Value)
	require.Equal(t, 2, scores[2].Value)
	require.Equal(t, 5, scores[3].Value)
	require.Equal(t, 18, scores[4].Value)
	require.Equal(t, 19, scores[5].Value)

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Reverse().Find(&scores)
	require.NoError(t, err)
	require.Len(t, scores, 6)
	require.Equal(t, 19, scores[0].Value)
	require.Equal(t, 18, scores[1].Value)
	require.Equal(t, 5, scores[2].Value)
	require.Equal(t, 2, scores[3].Value)
	require.Equal(t, 1, scores[4].Value)
	require.Equal(t, 0, scores[5].Value)
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
	require.NoError(t, err)
	require.Len(t, scores, 2)
	require.Equal(t, 18, scores[0].Value)
	require.Equal(t, 19, scores[1].Value)

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(-10).Find(&scores)
	require.NoError(t, err)
	require.Len(t, scores, 6)
	require.Equal(t, 0, scores[0].Value)

	scores = nil
	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(1000).Find(&scores)
	require.Error(t, err)
	require.True(t, ErrNotFound == err)
	require.Len(t, scores, 0)
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
	require.NoError(t, err)
	require.Len(t, scores, 2)
	require.Equal(t, 0, scores[0].Value)
	require.Equal(t, 1, scores[1].Value)

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Limit(-10).Find(&scores)
	require.NoError(t, err)
	require.Len(t, scores, 6)
	require.Equal(t, 0, scores[0].Value)

	scores = nil
	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Limit(0).Find(&scores)
	require.Error(t, err)
	require.True(t, ErrNotFound == err)
	require.Len(t, scores, 0)
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
	require.NoError(t, err)
	require.Len(t, scores, 2)
	require.Equal(t, 2, scores[0].Value)
	require.Equal(t, 5, scores[1].Value)

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Limit(2).Skip(5).Find(&scores)
	require.NoError(t, err)
	require.Len(t, scores, 1)
	require.Equal(t, 19, scores[0].Value)
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
		require.NoError(t, err)
	}

	var list []T
	err := db.Select().OrderBy("ID").Find(&list)
	require.NoError(t, err)
	require.Len(t, list, 6)
	for i, j := 0, 0; i < 6; i, j = i+1, j+1 {
		if i == 2 {
			j--
		}
		require.Equal(t, i+1, list[i].ID)
	}

	list = nil
	err = db.Select().OrderBy("Str").Find(&list)
	require.NoError(t, err)
	require.Len(t, list, 6)
	for i, j := 0, 0; i < 6; i, j = i+1, j+1 {
		if i == 4 {
			j--
		}
		require.Equal(t, string([]byte{'a' + byte(j)}), list[i].Str)
	}

	list = nil
	err = db.Select().OrderBy("Int").Find(&list)
	require.NoError(t, err)
	require.Len(t, list, 6)
	for i, j := 0, 0; i < 6; i, j = i+1, j+1 {
		if i == 2 {
			j--
		}
		require.Equal(t, j+1, list[i].Int)
	}

	list = nil
	err = db.Select().OrderBy("Rnd").Find(&list)
	require.NoError(t, err)
	require.Len(t, list, 6)
	require.Equal(t, 1, list[0].ID)
	require.Equal(t, 2, list[1].ID)
	require.Equal(t, 3, list[2].ID)
	require.Equal(t, 5, list[3].ID)
	require.Equal(t, 6, list[4].ID)
	require.Equal(t, 4, list[5].ID)

	list = nil
	err = db.Select().OrderBy("Int").Reverse().Find(&list)
	require.NoError(t, err)
	require.Len(t, list, 6)
	for i, j := 0, 0; i < 6; i, j = i+1, j+1 {
		if i == 4 {
			j--
		}
		require.Equal(t, 5-j, list[i].Int)
	}

	list = nil
	err = db.Select().OrderBy("Int").Reverse().Limit(2).Find(&list)
	require.NoError(t, err)
	require.Len(t, list, 2)
	for i := 0; i < 2; i++ {
		require.Equal(t, 5-i, list[i].Int)
	}

	list = nil
	err = db.Select().OrderBy("Int").Reverse().Skip(2).Find(&list)
	require.NoError(t, err)
	require.Len(t, list, 4)
	for i, j := 0, 0; i < 3; i, j = i+1, j+1 {
		if i == 2 {
			j--
		}
		require.Equal(t, 3-j, list[i].Int)
	}

	list = nil
	err = db.Select().OrderBy("Int").Reverse().Skip(5).Limit(2).Find(&list)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, 1, list[0].Int)

	list = nil
	err = db.Select().OrderBy("Str", "Int").Find(&list)
	require.NoError(t, err)
	require.Len(t, list, 6)
	require.Equal(t, "a", list[0].Str)
	require.Equal(t, 4, list[0].Int)
	require.Equal(t, "b", list[1].Str)
	require.Equal(t, 3, list[1].Int)
	require.Equal(t, "c", list[2].Str)
	require.Equal(t, 2, list[2].Int)
	require.Equal(t, "d", list[3].Str)
	require.Equal(t, 1, list[3].Int)
	require.Equal(t, "d", list[4].Str)
	require.Equal(t, 5, list[4].Int)
	require.Equal(t, "e", list[5].Str)
	require.Equal(t, 2, list[5].Int)
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
	require.NoError(t, err)
	require.Equal(t, 2, score.Value)

	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(1).Reverse().First(&score)
	require.NoError(t, err)
	require.Equal(t, 18, score.Value)
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
		require.NoError(t, err)
	}

	var record T
	err := db.Select().OrderBy("ID").First(&record)
	require.NoError(t, err)
	require.Equal(t, 1, record.ID)

	err = db.Select().OrderBy("Str").First(&record)
	require.NoError(t, err)
	require.Equal(t, "a", record.Str)

	err = db.Select().OrderBy("Int").First(&record)
	require.NoError(t, err)
	require.Equal(t, 1, record.Int)

	err = db.Select().OrderBy("Int").Reverse().First(&record)
	require.NoError(t, err)
	require.Equal(t, 5, record.Int)

	err = db.Select().OrderBy("Int").Reverse().Limit(2).First(&record)
	require.NoError(t, err)
	require.Equal(t, 5, record.Int)

	err = db.Select().OrderBy("Int").Reverse().Skip(2).First(&record)
	require.NoError(t, err)
	require.Equal(t, 3, record.Int)

	err = db.Select().OrderBy("Int").Reverse().Skip(4).Limit(2).First(&record)
	require.NoError(t, err)
	require.Equal(t, 1, record.Int)
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
	require.NoError(t, err)

	var scores []Score
	err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Find(&scores)
	require.NoError(t, err)
	require.Len(t, scores, 2)
	require.Equal(t, 0, scores[0].Value)
	require.Equal(t, 1, scores[1].Value)

	for i := 0; i < 10; i++ {
		w := User{ID: i + 1, Name: fmt.Sprintf("John%d", i+1)}
		err = db.Save(&w)
		require.NoError(t, err)
	}

	err = db.Select(q.Gte("ID", 5)).Delete(&User{})
	require.NoError(t, err)

	var user User
	err = db.One("Name", "John6", &user)
	require.Error(t, err)
	require.Equal(t, ErrNotFound, err)

	err = db.One("Name", "John4", &user)
	require.NoError(t, err)
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
	require.NoError(t, err)
	require.Equal(t, 6, total)

	total, err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(2).Count(&Score{})
	require.NoError(t, err)
	require.Equal(t, 4, total)

	total, err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(2).Limit(2).Count(&Score{})
	require.NoError(t, err)
	require.Equal(t, 2, total)

	total, err = db.Select(q.Or(
		q.Eq("Value", 5),
		q.Or(
			q.Lte("Value", 2),
			q.Gte("Value", 18),
		),
	)).Skip(5).Limit(2).Count(&Score{})
	require.NoError(t, err)
	require.Equal(t, 1, total)
}

func TestSelectRaw(t *testing.T) {
	db, cleanup := createDB(t, Codec(json.Codec))
	defer cleanup()

	for i := 0; i < 20; i++ {
		err := db.Save(&Score{
			Value: i,
		})
		require.NoError(t, err)
	}

	list, err := db.Select().Bucket("Score").Raw()
	require.NoError(t, err)
	require.Len(t, list, 20)

	list, err = db.Select().Bucket("Score").Skip(18).Limit(5).Raw()
	require.NoError(t, err)
	require.Len(t, list, 2)

	i := 0
	err = db.Select().Bucket("Score").Skip(18).Limit(5).RawEach(func(k []byte, v []byte) error {
		i++
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, i, 2)
}

func TestSelectEach(t *testing.T) {
	db, cleanup := createDB(t, Codec(json.Codec))
	defer cleanup()

	for i := 0; i < 20; i++ {
		err := db.Save(&Score{
			Value: i,
		})
		require.NoError(t, err)
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
