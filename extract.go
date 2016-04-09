package storm

import (
	"fmt"

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
	ID      *structs.Field
	Indexes map[string]indexInfo
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
	}

	if m.Name == "" {
		m.Name = s.Name()
	}

	for _, f := range fields {
		if !f.IsExported() {
			continue
		}

		tag := f.Tag("storm")
		if tag != "" {
			switch tag {
			case "id":
				m.ID = f
			case tagUniqueIdx, tagIdx:
				m.AddIndex(f, tag, !child)
			case tagInline:
				if structs.IsStruct(f.Value()) {
					_, err := extract(f.Value(), m)
					if err != nil {
						return nil, err
					}
				}
			default:
				return nil, fmt.Errorf("unknown tag %s", tag)
			}
		}

		// the field is named ID and no ID field has been detected before
		if f.Name() == "ID" && m.ID == nil {
			m.ID = f
		}
	}

	// ID field or tag detected, add to the unique index
	if m.ID != nil {
		m.AddIndex(m.ID, tagUniqueIdx, !child)
	}

	return m, nil
}
