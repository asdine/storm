package storm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	err := db.Set("trash", 10, 100)
	assert.NoError(t, err)

	var nb int
	err = db.Get("trash", 10, &nb)
	assert.NoError(t, err)
	assert.Equal(t, 100, nb)

	tm := time.Now()
	err = db.Set("logs", tm, "I'm hungry")
	assert.NoError(t, err)

	var message string
	err = db.Get("logs", tm, &message)
	assert.NoError(t, err)
	assert.Equal(t, "I'm hungry", message)

	var hand int
	err = db.Get("wallet", "100 bucks", &hand)
	assert.EqualError(t, err, "not found")

	err = db.Set("wallet", "10 bucks", 10)
	assert.NoError(t, err)

	err = db.Get("wallet", "100 bucks", &hand)
	assert.EqualError(t, err, "not found")

	err = db.Get("logs", tm, nil)
	assert.EqualError(t, err, "provided target must be a pointer to a valid variable")
}

func TestOneByIndex(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"))

	u := UniqueNameUser{Name: "John", ID: 10}
	err := db.Save(&u)
	assert.NoError(t, err)

	v := UniqueNameUser{}
	err = db.OneByIndex("Name", "John", &v)
	assert.NoError(t, err)
	assert.Equal(t, u, v)

	for i := 0; i < 10; i++ {
		w := IndexedNameUser{Name: "John", ID: i + 1}
		err := db.Save(&w)
		assert.NoError(t, err)
	}

	x := IndexedNameUser{}
	err = db.OneByIndex("Name", "John", &x)
	assert.NoError(t, err)
	assert.Equal(t, IndexedNameUser{Name: "John", ID: 1}, x)

	err = db.OneByIndex("Name", "Mike", &x)
	assert.Error(t, err)
	assert.EqualError(t, err, "not found")
}
