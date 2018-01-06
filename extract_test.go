package storm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.Error(t, err)
	require.Equal(t, ErrNoID, err)
}

func TestExtractBadTags(t *testing.T) {
	s := ClassicBadTags{}
	r := reflect.ValueOf(&s)
	infos, err := extract(&r)
	require.Error(t, err)
	require.Equal(t, ErrUnknownTag, err)
	require.Nil(t, infos)
}

func TestExtractUniqueTags(t *testing.T) {
	s := ClassicUnique{ID: "id"}
	r := reflect.ValueOf(&s)
	infos, err := extract(&r)
	require.NoError(t, err)
	require.NotNil(t, infos)
	require.NotNil(t, infos.ID)
	require.False(t, infos.ID.IsZero)
	require.Equal(t, "ClassicUnique", infos.Name)
	require.Len(t, allByType(infos, "index"), 0)
	require.Len(t, allByType(infos, "unique"), 5)
}

func TestExtractIndexTags(t *testing.T) {
	s := ClassicIndex{ID: "id"}
	r := reflect.ValueOf(&s)
	infos, err := extract(&r)
	require.NoError(t, err)
	require.NotNil(t, infos)
	require.NotNil(t, infos.ID)
	require.False(t, infos.ID.IsZero)
	require.Equal(t, "ClassicIndex", infos.Name)
	require.Len(t, allByType(infos, "index"), 5)
	require.Len(t, allByType(infos, "unique"), 1)
}

func TestExtractInlineWithIndex(t *testing.T) {
	s := ClassicInline{ToEmbed: &ToEmbed{ID: "50"}}
	r := reflect.ValueOf(&s)
	infos, err := extract(&r)
	require.NoError(t, err)
	require.NotNil(t, infos)
	require.NotNil(t, infos.ID)
	require.Equal(t, "ClassicInline", infos.Name)
	require.Len(t, allByType(infos, "index"), 3)
	require.Len(t, allByType(infos, "unique"), 3)
}

func TestExtractMultipleTags(t *testing.T) {
	type User struct {
		ID              uint64 `storm:"id,increment"`
		Age             uint16 `storm:"index,increment"`
		unexportedField int32  `storm:"index,increment"`
		X               uint32 `storm:"unique,increment=100"`
		Y               int8   `storm:"index,increment=-100"`
	}

	s := User{}
	r := reflect.ValueOf(&s)
	infos, err := extract(&r)
	require.NoError(t, err)
	require.NotNil(t, infos)
	require.NotNil(t, infos.ID)
	require.Equal(t, "User", infos.Name)
	require.Len(t, allByType(infos, "index"), 2)
	require.Len(t, allByType(infos, "unique"), 2)

	require.True(t, infos.Fields["Age"].Increment)
	require.Equal(t, int64(1), infos.Fields["Age"].IncrementStart)
	require.Equal(t, "index", infos.Fields["Age"].Index)
	require.False(t, infos.Fields["Age"].IsID)
	require.True(t, infos.Fields["Age"].IsInteger)
	require.True(t, infos.Fields["Age"].IsZero)
	require.NotNil(t, infos.Fields["Age"].Value)

	require.True(t, infos.Fields["X"].Increment)
	require.Equal(t, int64(100), infos.Fields["X"].IncrementStart)
	require.Equal(t, "unique", infos.Fields["X"].Index)
	require.False(t, infos.Fields["X"].IsID)
	require.True(t, infos.Fields["X"].IsInteger)
	require.True(t, infos.Fields["X"].IsZero)
	require.NotNil(t, infos.Fields["X"].Value)

	require.True(t, infos.Fields["Y"].Increment)
	require.Equal(t, int64(-100), infos.Fields["Y"].IncrementStart)
	require.Equal(t, "index", infos.Fields["Y"].Index)
	require.False(t, infos.Fields["Y"].IsID)
	require.True(t, infos.Fields["Y"].IsInteger)
	require.True(t, infos.Fields["Y"].IsZero)
	require.NotNil(t, infos.Fields["Y"].Value)

	type NoInt struct {
		ID uint64 `storm:"id,increment=hello"`
	}

	var n NoInt
	r = reflect.ValueOf(&n)
	_, err = extract(&r)
	require.Error(t, err)

	type BadSuffix struct {
		ID uint64 `storm:"id,incrementag=100"`
	}

	var b BadSuffix
	r = reflect.ValueOf(&b)
	_, err = extract(&r)
	require.Error(t, err)
}
