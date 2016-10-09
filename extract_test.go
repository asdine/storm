package storm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func allByType(m *structConfig, indexType string) []*fieldConfig {
	var idx []*fieldConfig
	for k := range m.Fields {
		if m.Fields[k].Index == indexType {
			idx = append(idx, m.Fields[k])
		}
	}

	return idx
}

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
	assert.Len(t, allByType(infos, "index"), 0)
	assert.Len(t, allByType(infos, "unique"), 4)
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
	assert.Len(t, allByType(infos, "index"), 5)
	assert.Len(t, allByType(infos, "unique"), 0)
}

func TestExtractInlineWithIndex(t *testing.T) {
	s := ClassicInline{ToEmbed: &ToEmbed{ID: "50"}}
	r := reflect.ValueOf(&s)
	infos, err := extract(&r)
	assert.NoError(t, err)
	assert.NotNil(t, infos)
	assert.NotNil(t, infos.ID)
	assert.Equal(t, "ClassicInline", infos.Name)
	assert.Len(t, allByType(infos, "index"), 3)
	assert.Len(t, allByType(infos, "unique"), 2)
}

func TestExtractMultipleTags(t *testing.T) {
	type User struct {
		ID              uint64 `storm:"id,increment"`
		Age             uint16 `storm:"index,increment"`
		unexportedField int32  `storm:"index,increment"`
		Pos             string `storm:"unique,increment"`
	}

	s := User{}
	r := reflect.ValueOf(&s)
	infos, err := extract(&r)
	assert.NoError(t, err)
	assert.NotNil(t, infos)
	assert.NotNil(t, infos.ID)
	assert.Equal(t, "User", infos.Name)
	assert.Len(t, allByType(infos, "index"), 1)
	assert.Len(t, allByType(infos, "unique"), 1)
	assert.True(t, infos.Fields["Age"].Increment)
	assert.Equal(t, "index", infos.Fields["Age"].Index)
	assert.False(t, infos.Fields["Age"].IsID)
	assert.True(t, infos.Fields["Age"].IsInteger)
	assert.True(t, infos.Fields["Age"].IsZero)
	assert.NotNil(t, infos.Fields["Age"].Value)
}
