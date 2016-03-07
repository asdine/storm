package storm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractNoTags(t *testing.T) {
	s := ClassicNoTags{}
	infos, err := extract(&s)
	assert.NoError(t, err)
	assert.NotNil(t, infos)
	assert.Nil(t, infos.ID)
	assert.Equal(t, "ClassicNoTags", infos.Name)
	assert.Len(t, infos.AllByType("index"), 0)
}

func TestExtractBadTags(t *testing.T) {
	s := ClassicBadTags{}
	infos, err := extract(&s)
	assert.Error(t, err)
	assert.EqualError(t, err, "unknown tag mrots")
	assert.Nil(t, infos)
}

func TestExtractUniqueTags(t *testing.T) {
	s := ClassicUnique{}
	infos, err := extract(&s)
	assert.NoError(t, err)
	assert.NotNil(t, infos)
	assert.Nil(t, infos.ID)
	assert.Equal(t, "ClassicUnique", infos.Name)
	assert.Len(t, infos.AllByType("index"), 0)
	assert.Len(t, infos.AllByType("unique"), 4)
}

func TestExtractIndexTags(t *testing.T) {
	s := ClassicIndex{}
	infos, err := extract(&s)
	assert.NoError(t, err)
	assert.NotNil(t, infos)
	assert.Nil(t, infos.ID)
	assert.Equal(t, "ClassicIndex", infos.Name)
	assert.Len(t, infos.AllByType("index"), 5)
	assert.Len(t, infos.AllByType("unique"), 0)
}

func TestExtractInlineWithIndex(t *testing.T) {
	s := ClassicInline{ToEmbed: &ToEmbed{ID: "50"}}
	infos, err := extract(&s)
	assert.NoError(t, err)
	assert.NotNil(t, infos)
	assert.NotNil(t, infos.ID)
	assert.Equal(t, "ClassicInline", infos.Name)
	assert.Len(t, infos.AllByType("index"), 3)
	assert.Len(t, infos.AllByType("unique"), 3)
}
