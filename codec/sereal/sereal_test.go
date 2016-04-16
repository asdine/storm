package sereal

import (
	"testing"

	"github.com/asdine/storm/codec"
	"github.com/stretchr/testify/assert"
)

type SerealUser struct {
	Name string
	Self *SerealUser
}

func TestSereal(t *testing.T) {
	u1 := &SerealUser{Name: "Sereal"}
	u1.Self = u1 // cyclic ref
	u2 := &SerealUser{}
	codec.RountripTester(t, Codec, &u1, &u2)
	assert.True(t, u2 == u2.Self)
}
