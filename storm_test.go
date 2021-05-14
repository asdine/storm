package storm_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/v4"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	dir, err := os.MkdirTemp("", "storm")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	tests := []struct {
		path  string
		fails bool
	}{
		{filepath.Join(dir, "foo.db"), false},
		{filepath.Join(dir, "foo"), false},
		{"/doesntexist/doesntexisteither", true},
		{":memory:", false},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			db, err := storm.Open(test.path)
			if test.fails {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				err = db.Close()
				require.NoError(t, err)
			}
		})
	}
}
