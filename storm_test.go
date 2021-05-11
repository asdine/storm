package storm_test

import (
	"testing"

	"github.com/asdine/storm/v4"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	db, err := storm.Open(":memory:")
	require.NoError(t, err)
	defer db.Close()
}
