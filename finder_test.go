package storm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestFindNil(t *testing.T) {
	db, cleanup := createDB(t, AutoIncrement())
	defer cleanup()

	type User struct {
		ID        int        `storm:"increment"`
		CreatedAt *time.Time `storm:"index"`
		DeletedAt *time.Time `storm:"unique"`
	}

	t1 := time.Now()
	for i := 0; i < 10; i++ {
		now := time.Now()
		var u User

		if i%2 == 0 {
			u.CreatedAt = &t1
			u.DeletedAt = &now
		}

		err := db.Save(&u)
		assert.NoError(t, err)
	}

	var users []User
	err := db.Find("CreatedAt", nil, &users)
	require.NoError(t, err)
	require.Len(t, users, 5)

	users = nil
	err = db.Find("CreatedAt", t1, &users)
	require.NoError(t, err)
	require.Len(t, users, 5)

	users = nil
	err = db.Find("DeletedAt", nil, &users)
	require.NoError(t, err)
	require.Len(t, users, 5)
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

	_, err = tx.Count(User{})
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
