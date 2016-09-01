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
)

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

	x := IndexedNameUser{}
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

func BenchmarkOneWithIndex(b *testing.B) {
	db, cleanup := createDB(b, AutoIncrement())
	defer cleanup()

	var u User
	for i := 0; i < 100; i++ {
		w := User{Name: fmt.Sprintf("John%d", i), Group: fmt.Sprintf("Staff%d", i)}
		err := db.Save(&w)
		if err != nil {
			b.Error(err)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := db.One("Name", "John99", &u)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkOneByID(b *testing.B) {
	db, cleanup := createDB(b, AutoIncrement())
	defer cleanup()

	type User struct {
		ID          int
		Name        string `storm:"index"`
		age         int
		DateOfBirth time.Time `storm:"index"`
		Group       string
		Slug        string `storm:"unique"`
	}

	var u User
	for i := 0; i < 100; i++ {
		w := User{Name: fmt.Sprintf("John%d", i), Group: fmt.Sprintf("Staff%d", i)}
		err := db.Save(&w)
		if err != nil {
			b.Error(err)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := db.One("ID", 99, &u)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkOneWithoutIndex(b *testing.B) {
	db, cleanup := createDB(b, AutoIncrement())
	defer cleanup()

	var u User
	for i := 0; i < 100; i++ {
		w := User{Name: "John", Group: fmt.Sprintf("Staff%d", i)}
		err := db.Save(&w)
		if err != nil {
			b.Error(err)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := db.One("Group", "Staff99", &u)
		if err != nil {
			b.Error(err)
		}
	}
}
