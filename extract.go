package storm

import (
	"reflect"

	"github.com/fatih/structs"
)

// Storm tags
const (
	tagID        = "id"
	tagIdx       = "index"
	tagUniqueIdx = "unique"
	tagInline    = "inline"
)

type indexInfo struct {
	Type  string
	Field *structs.Field
}

// modelInfo is a structure gathering all the relevant informations about a model
type modelInfo struct {
	Name    string
	Indexes map[string]indexInfo
	ID      identInfo
	data    interface{}
}

func (m *modelInfo) AddIndex(f *structs.Field, indexType string, override bool) {
	fieldName := f.Name()
	if _, ok := m.Indexes[fieldName]; !ok || override {
		m.Indexes[fieldName] = indexInfo{
			Type:  indexType,
			Field: f,
		}
	}
}

func (m *modelInfo) AllByType(indexType string) []indexInfo {
	var idx []indexInfo
	for k := range m.Indexes {
		if m.Indexes[k].Type == indexType {
			idx = append(idx, m.Indexes[k])
		}
	}

	return idx
}

func extract(data interface{}, mi ...*modelInfo) (*modelInfo, error) {
	s := structs.New(data)
	fields := s.Fields()

	var child bool

	var m *modelInfo
	if len(mi) > 0 {
		m = mi[0]
		child = true
	} else {
		m = &modelInfo{}
		m.Indexes = make(map[string]indexInfo)
		m.data = data
	}

	if m.Name == "" {
		m.Name = s.Name()
	}

	for _, f := range fields {
		if !f.IsExported() {
			continue
		}

		err := extractField(f, m, child)
		if err != nil {
			return nil, err
		}
	}

	// ID field or tag detected
	if m.ID.Field != nil {
		if m.ID.Field.IsZero() {
			m.ID.IsZero = true
		} else {
			m.ID.Value = m.ID.Field.Value()
		}
	}

	if child {
		return m, nil
	}

	if m.ID.Field == nil {
		return nil, ErrNoID
	}

	if m.Name == "" {
		return nil, ErrNoName
	}

	return m, nil
}

func extractField(f *structs.Field, m *modelInfo, isChild bool) error {
	tag := f.Tag("storm")
	if tag != "" {
		switch tag {
		case "id":
			m.ID.Field = f
		case tagUniqueIdx, tagIdx:
			m.AddIndex(f, tag, !isChild)
		case tagInline:
			if structs.IsStruct(f.Value()) {
				_, err := extract(f.Value(), m)
				if err != nil {
					return err
				}
			}
		default:
			return ErrUnknownTag
		}
	}

	// the field is named ID and no ID field has been detected before
	if f.Name() == "ID" && m.ID.Field == nil {
		m.ID.Field = f
	}

	return nil
}

// Prefill the most requested informations
type identInfo struct {
	Field  *structs.Field
	IsZero bool
	Value  interface{}
}

func (i *identInfo) Type() reflect.Type {
	return reflect.TypeOf(i.Field.Value())
}

func (i *identInfo) IsOfIntegerFamily() bool {
	return i.Field != nil && i.Field.Kind() >= reflect.Int && i.Field.Kind() <= reflect.Uint64
}
