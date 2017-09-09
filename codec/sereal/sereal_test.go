package sereal

import (
	"testing"

	"github.com/asdine/storm/codec/internal"
	"github.com/stretchr/testify/require"
)

type SerealUser struct {
	Name string
	Self *SerealUser
}

func TestSereal(t *testing.T) {
	u1 := &SerealUser{Name: "Sereal"}
	u1.Self = u1 // cyclic ref
	u2 := &SerealUser{}
	internal.RoundtripTester(t, Codec, &u1, &u2)
	require.True(t, u2 == u2.Self)
}
