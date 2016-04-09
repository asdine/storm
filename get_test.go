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
	assert.Equal(t, ErrNotFound, err)

	err = db.Set("wallet", "10 bucks", 10)
	assert.NoError(t, err)

	err = db.Get("wallet", "100 bucks", &hand)
	assert.Equal(t, ErrNotFound, err)

	err = db.Get("logs", tm, nil)
	assert.Equal(t, ErrPtrNeeded, err)

	err = db.Get("", nil, nil)
	assert.Equal(t, ErrPtrNeeded, err)

	err = db.Get("", "100 bucks", &hand)
	assert.Equal(t, ErrNotFound, err)
}
