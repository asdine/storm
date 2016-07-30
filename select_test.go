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

func TestSelect(t *testing.T) {
	dir, _ := ioutil.TempDir(os.TempDir(), "storm")
	defer os.RemoveAll(dir)
	db, _ := Open(filepath.Join(dir, "storm.db"), AutoIncrement())
	defer db.Close()

	for i := 0; i < 10; i++ {
		err := db.Save(&Score{
			Value: i,
		})
		assert.NoError(t, err)
	}

	var scores []Score

	err := db.Select(&scores, q.Eq("Value", 5))
	assert.NoError(t, err)
	assert.Len(t, scores, 1)
	assert.Equal(t, 5, scores[0].Value)

	err = db.Select(&scores,
		q.Or(
			q.Eq("Value", 5),
			q.Eq("Value", 6),
		),
	)
	assert.NoError(t, err)
	assert.Len(t, scores, 2)
	assert.Equal(t, 5, scores[0].Value)
	assert.Equal(t, 6, scores[1].Value)
}
