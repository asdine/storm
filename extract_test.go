package storm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractNoTags(t *testing.T) {
	s := ClassicNoTags{}
	r := reflect.ValueOf(&s)
	_, err := extract(&r)
	assert.Error(t, err)
	assert.Equal(t, ErrNoID, err)
}

func TestExtractBadTags(t *testing.T) {
	s := ClassicBadTags{}
	r := reflect.ValueOf(&s)
	infos, err := extract(&r)
	assert.Error(t, err)
	assert.Equal(t, ErrUnknownTag, err)
	assert.Nil(t, infos)
}

func TestExtractUniqueTags(t *testing.T) {
	s := ClassicUnique{ID: "id"}
	r := reflect.ValueOf(&s)
	infos, err := extract(&r)
	assert.NoError(t, err)
	assert.NotNil(t, infos)
	assert.NotNil(t, infos.ID)
	assert.False(t, infos.ID.IsZero)
	assert.Equal(t, "ClassicUnique", infos.Name)
	assert.Len(t, infos.AllByType("index"), 0)
	assert.Len(t, infos.AllByType("unique"), 4)
}

func TestExtractIndexTags(t *testing.T) {
	s := ClassicIndex{ID: "id"}
	r := reflect.ValueOf(&s)
	infos, err := extract(&r)
	assert.NoError(t, err)
	assert.NotNil(t, infos)
	assert.NotNil(t, infos.ID)
	assert.False(t, infos.ID.IsZero)
	assert.Equal(t, "ClassicIndex", infos.Name)
	assert.Len(t, infos.AllByType("index"), 5)
	assert.Len(t, infos.AllByType("unique"), 0)
}

func TestExtractInlineWithIndex(t *testing.T) {
	s := ClassicInline{ToEmbed: &ToEmbed{ID: "50"}}
	r := reflect.ValueOf(&s)
	infos, err := extract(&r)
	assert.NoError(t, err)
	assert.NotNil(t, infos)
	assert.NotNil(t, infos.ID)
	assert.Equal(t, "ClassicInline", infos.Name)
	assert.Len(t, infos.AllByType("index"), 3)
	assert.Len(t, infos.AllByType("unique"), 2)
}
