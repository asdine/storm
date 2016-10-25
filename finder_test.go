package storm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/asdine/storm/codec/json"
	"github.com/asdine/storm/q"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFind(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	for i := 0; i < 100; i++ {
		w := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%d", i+1)}

		if i%2 == 0 {
			w.Group = "staff"
		} else {
			w.Group = "normal"
		}

		err := db.Save(&w)
		assert.NoError(t, err)
	}

	err := db.Find("Name", "John", &User{})
	assert.Error(t, err)
	assert.Equal(t, ErrSlicePtrNeeded, err)

	err = db.Find("Name", "John", &[]struct {
		Name string
		ID   int
	}{})
	assert.Error(t, err)
	assert.Equal(t, ErrNoName, err)

	notTheRightUsers := []UniqueNameUser{}

	err = db.Find("Name", "John", &notTheRightUsers)
	assert.Error(t, err)
	assert.EqualError(t, err, "not found")

	users := []User{}

	err = db.Find("unexportedField", "John", &users)
	assert.Error(t, err)
	assert.EqualError(t, err, "field unexportedField not found")

	err = db.Find("DateOfBirth", "John", &users)
	assert.Error(t, err)
	assert.True(t, ErrNotFound == err)

	err = db.Find("Group", "staff", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 50)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 99, users[49].ID)

	err = db.Find("Group", "staff", &users, Reverse())
	assert.NoError(t, err)
	assert.Len(t, users, 50)
	assert.Equal(t, 99, users[0].ID)
	assert.Equal(t, 1, users[49].ID)

	err = db.Find("Group", "admin", &users)
	assert.Error(t, err)
	assert.True(t, ErrNotFound == err)

	err = db.Find("Name", "John", users)
	assert.Error(t, err)
	assert.Equal(t, ErrSlicePtrNeeded, err)

	err = db.Find("Name", "John", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	err = db.Find("Name", "John", &users, Reverse())
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 100, users[0].ID)
	assert.Equal(t, 1, users[99].ID)

	users = []User{}
	err = db.Find("Slug", "John10", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, 10, users[0].ID)

	users = []User{}
	err = db.Find("Name", nil, &users)
	assert.Error(t, err)
	assert.True(t, ErrNotFound == err)

	err = db.Find("Name", "John", &users, Limit(10), Skip(20))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 21, users[0].ID)
	assert.Equal(t, 30, users[9].ID)

	err = db.Find("Age", 10, &users)
	assert.NoError(t, err)
}

func TestFindIntIndex(t *testing.T) {
	db, cleanup := createDB(t, AutoIncrement())
	defer cleanup()

	type Score struct {
		ID    int
		Score uint64 `storm:"index"`
	}

	for i := 0; i < 10; i++ {
		w := Score{Score: uint64(i % 3)}
		err := db.Save(&w)
		require.NoError(t, err)
	}

	var scores []Score
	err := db.Find("Score", 2, &scores)
	require.NoError(t, err)
	require.Len(t, scores, 3)
	require.Equal(t, []Score{
		{ID: 3, Score: 2},
		{ID: 6, Score: 2},
		{ID: 9, Score: 2},
	}, scores)
}

func TestAllByIndex(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	for i := 0; i < 100; i++ {
		w := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%d", i+1), DateOfBirth: time.Now().Add(-time.Duration(i*10) * time.Minute)}
		err := db.Save(&w)
		assert.NoError(t, err)
	}

	err := db.AllByIndex("", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrSlicePtrNeeded, err)

	var users []User

	err = db.AllByIndex("Unknown field", &users)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	err = db.AllByIndex("DateOfBirth", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 100, users[0].ID)
	assert.Equal(t, 1, users[99].ID)

	err = db.AllByIndex("Name", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	y := UniqueNameUser{Name: "Jake", ID: 200}
	err = db.Save(&y)
	assert.NoError(t, err)

	var y2 []UniqueNameUser
	err = db.AllByIndex("ID", &y2)
	assert.NoError(t, err)
	assert.Len(t, y2, 1)

	n := NestedID{}
	n.ID = "100"
	n.Name = "John"

	err = db.Save(&n)
	assert.NoError(t, err)

	var n2 []NestedID
	err = db.AllByIndex("ID", &n2)
	assert.NoError(t, err)
	assert.Len(t, n2, 1)

	err = db.AllByIndex("Name", &users, Limit(10))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 10, users[9].ID)

	err = db.AllByIndex("Name", &users, Limit(200))
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	err = db.AllByIndex("Name", &users, Limit(-10))
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	err = db.AllByIndex("Name", &users, Skip(200))
	assert.NoError(t, err)
	assert.Len(t, users, 0)

	err = db.AllByIndex("Name", &users, Skip(-10))
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	err = db.AllByIndex("ID", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	err = db.AllByIndex("ID", &users, Limit(10))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 10, users[9].ID)

	err = db.AllByIndex("ID", &users, Skip(10))
	assert.NoError(t, err)
	assert.Len(t, users, 90)
	assert.Equal(t, 11, users[0].ID)
	assert.Equal(t, 100, users[89].ID)

	err = db.AllByIndex("Name", &users, Limit(10), Skip(10))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 11, users[0].ID)
	assert.Equal(t, 20, users[9].ID)

	err = db.AllByIndex("Name", &users, Limit(10), Skip(10), Reverse())
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 90, users[0].ID)
	assert.Equal(t, 81, users[9].ID)

	err = db.AllByIndex("Age", &users, Limit(10))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
}

func TestAll(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	for i := 0; i < 100; i++ {
		w := User{Name: "John", ID: i + 1, Slug: fmt.Sprintf("John%d", i+1), DateOfBirth: time.Now().Add(-time.Duration(i*10) * time.Minute)}
		err := db.Save(&w)
		assert.NoError(t, err)
	}

	var users []User

	err := db.All(&users)
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, 100, users[99].ID)

	err = db.All(&users, Reverse())
	assert.NoError(t, err)
	assert.Len(t, users, 100)
	assert.Equal(t, 100, users[0].ID)
	assert.Equal(t, 1, users[99].ID)

	var users2 []*User

	err = db.All(&users2)
	assert.NoError(t, err)
	assert.Len(t, users2, 100)
	assert.Equal(t, 1, users2[0].ID)
	assert.Equal(t, 100, users2[99].ID)

	err = db.Save(&NestedID{
		ToEmbed: ToEmbed{ID: "id1"},
		Name:    "John",
	})
	assert.NoError(t, err)

	err = db.Save(&NestedID{
		ToEmbed: ToEmbed{ID: "id2"},
		Name:    "Mike",
	})
	assert.NoError(t, err)

	db.Save(&NestedID{
		ToEmbed: ToEmbed{ID: "id3"},
		Name:    "Steve",
	})
	assert.NoError(t, err)

	var nested []NestedID
	err = db.All(&nested)
	assert.NoError(t, err)
	assert.Len(t, nested, 3)

	err = db.All(&users, Limit(10), Skip(10))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 11, users[0].ID)
	assert.Equal(t, 20, users[9].ID)

	err = db.All(&users, Limit(0), Skip(0))
	assert.NoError(t, err)
	assert.Len(t, users, 0)
}

func TestCount(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

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

func TestCountEmpty(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	user := &User{}
	err := db.Init(user)
	assert.NoError(t, err)

	count, err := db.Count(user)
	assert.Zero(t, count)
	assert.NoError(t, err)
}

func TestOne(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	u := UniqueNameUser{Name: "John", ID: 10}
	err := db.Save(&u)
	assert.NoError(t, err)

	v := UniqueNameUser{}
	err = db.One("Name", "John", &v)
	assert.NoError(t, err)
	assert.Equal(t, u, v)

	for i := 0; i < 10; i++ {
		w := IndexedNameUser{Name: "John", ID: i + 1, Group: "staff"}
		err = db.Save(&w)
		assert.NoError(t, err)
	}

	var x IndexedNameUser
	err = db.One("Name", "John", &x)
	assert.NoError(t, err)
	assert.Equal(t, "John", x.Name)
	assert.Equal(t, 1, x.ID)
	assert.Zero(t, x.age)
	assert.True(t, x.DateOfBirth.IsZero())

	err = db.One("Name", "Mike", &x)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	err = db.One("", nil, &x)
	assert.Error(t, err)
	assert.True(t, ErrNotFound == err)

	err = db.One("", "Mike", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrStructPtrNeeded, err)

	err = db.One("", nil, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrStructPtrNeeded, err)

	err = db.One("Group", "staff", &x)
	assert.NoError(t, err)
	assert.Equal(t, 1, x.ID)

	err = db.One("Score", 5, &x)
	assert.NoError(t, err)
	assert.Equal(t, 5, x.ID)

	err = db.One("Group", "admin", &x)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	y := UniqueNameUser{Name: "Jake", ID: 200}
	err = db.Save(&y)
	assert.NoError(t, err)

	var y2 UniqueNameUser
	err = db.One("ID", 200, &y2)
	assert.NoError(t, err)
	assert.Equal(t, y, y2)

	n := NestedID{}
	n.ID = "100"
	n.Name = "John"

	err = db.Save(&n)
	assert.NoError(t, err)

	var n2 NestedID
	err = db.One("ID", "100", &n2)
	assert.NoError(t, err)
	assert.Equal(t, n, n2)
}

func TestOneNotWritable(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	err := db.Save(&User{ID: 10, Name: "John"})
	assert.NoError(t, err)

	db.Close()

	db, _ = Open(filepath.Join(dir, "storm.db"), BoltOptions(0660, &bolt.Options{
		ReadOnly: true,
	}))
	defer db.Close()

	err = db.Save(&User{ID: 20, Name: "John"})
	assert.Error(t, err)

	var u User
	err = db.One("ID", 10, &u)
	assert.NoError(t, err)
	assert.Equal(t, 10, u.ID)
	assert.Equal(t, "John", u.Name)

	err = db.One("Name", "John", &u)
	assert.NoError(t, err)
	assert.Equal(t, 10, u.ID)
	assert.Equal(t, "John", u.Name)
}

func TestRange(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()

	for i := 0; i < 100; i++ {
		w := User{
			Name:        "John",
			ID:          i + 1,
			Slug:        fmt.Sprintf("John%03d", i+1),
			DateOfBirth: time.Now().Add(-time.Duration(i) * time.Hour),
			Group:       fmt.Sprintf("Group%03d", i%5),
		}
		err := db.Save(&w)
		assert.NoError(t, err)
		z := User{Name: fmt.Sprintf("Zach%03d", i+1), ID: i + 101, Slug: fmt.Sprintf("Zach%03d", i+1)}
		err = db.Save(&z)
		assert.NoError(t, err)
	}

	min := "John010"
	max := "John020"
	var users []User

	err := db.Range("Slug", min, max, users)
	assert.Equal(t, ErrSlicePtrNeeded, err)

	err = db.Range("Slug", min, max, &users)
	assert.NoError(t, err)
	assert.Len(t, users, 11)
	assert.Equal(t, "John010", users[0].Slug)
	assert.Equal(t, "John020", users[10].Slug)

	err = db.Range("Slug", min, max, &users, Reverse())
	assert.NoError(t, err)
	assert.Len(t, users, 11)
	assert.Equal(t, "John020", users[0].Slug)
	assert.Equal(t, "John010", users[10].Slug)

	min = "Zach010"
	max = "Zach020"
	users = nil
	err = db.Range("Name", min, max, &users)
	assert.NoError(t, err)
	assert.Len(t, users, 11)
	assert.Equal(t, "Zach010", users[0].Name)
	assert.Equal(t, "Zach020", users[10].Name)

	err = db.Range("Name", min, max, &users, Reverse())
	assert.NoError(t, err)
	assert.Len(t, users, 11)
	assert.Equal(t, "Zach020", users[0].Name)
	assert.Equal(t, "Zach010", users[10].Name)

	err = db.Range("Name", min, max, &User{})
	assert.Error(t, err)
	assert.Equal(t, ErrSlicePtrNeeded, err)

	notTheRightUsers := []UniqueNameUser{}

	err = db.Range("Name", min, max, &notTheRightUsers)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(notTheRightUsers))

	users = nil

	err = db.Range("Age", min, max, &users)
	assert.Error(t, err)
	assert.EqualError(t, err, "not found")

	err = db.Range("Age", 2, 5, &users)
	assert.NoError(t, err)
	assert.Len(t, users, 4)

	dateMin := time.Now().Add(-time.Duration(50) * time.Hour)
	dateMax := dateMin.Add(time.Duration(3) * time.Hour)
	err = db.Range("DateOfBirth", dateMin, dateMax, &users)
	assert.NoError(t, err)
	assert.Len(t, users, 3)
	assert.Equal(t, "John050", users[0].Slug)
	assert.Equal(t, "John048", users[2].Slug)

	err = db.Range("Slug", "John010", "John040", &users, Limit(10), Skip(20))
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 30, users[0].ID)
	assert.Equal(t, 39, users[9].ID)

	err = db.Range("Slug", "John010", "John040", &users, Limit(10), Skip(20), Reverse())
	assert.NoError(t, err)
	assert.Len(t, users, 10)
	assert.Equal(t, 20, users[0].ID)
	assert.Equal(t, 11, users[9].ID)

	err = db.Range("Group", "Group002", "Group004", &users)
	assert.NoError(t, err)
	assert.Len(t, users, 60)
}

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

	var list []T
	err := db.Select().OrderBy("ID").Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 5)
	for i := 0; i < 5; i++ {
		assert.Equal(t, i+1, list[i].ID)
	}

	err = db.Select().OrderBy("Str").Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 5)
	for i := 0; i < 5; i++ {
		assert.Equal(t, string([]byte{'a' + byte(i)}), list[i].Str)
	}

	err = db.Select().OrderBy("Int").Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 5)
	for i := 0; i < 5; i++ {
		assert.Equal(t, i+1, list[i].Int)
	}

	err = db.Select().OrderBy("Int").Reverse().Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 5)
	for i := 0; i < 5; i++ {
		assert.Equal(t, 5-i, list[i].Int)
	}

	err = db.Select().OrderBy("Int").Reverse().Limit(2).Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	for i := 0; i < 2; i++ {
		assert.Equal(t, 5-i, list[i].Int)
	}

	err = db.Select().OrderBy("Int").Reverse().Skip(2).Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 3)
	for i := 0; i < 2; i++ {
		assert.Equal(t, 3-i, list[i].Int)
	}

	err = db.Select().OrderBy("Int").Reverse().Skip(4).Limit(2).Find(&list)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, 1, list[0].Int)
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
