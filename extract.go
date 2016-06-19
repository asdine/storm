package storm

import (
	"reflect"

	"github.com/asdine/storm/index"
	"github.com/boltdb/bolt"
)

// Storm tags
const (
	tagID        = "id"
	tagIdx       = "index"
	tagUniqueIdx = "unique"
	tagInline    = "inline"
	indexPrefix  = "__storm_index_"
)

type indexInfo struct {
	Type  string
	Field *reflect.StructField
	Value *reflect.Value
}

func (i *indexInfo) IsZero() bool {
	zero := reflect.Zero(i.Value.Type()).Interface()
	current := i.Value.Interface()
	return reflect.DeepEqual(current, zero)
}

// modelInfo is a structure gathering all the relevant informations about a model
type modelInfo struct {
	Name    string
	Indexes map[string]indexInfo
	ID      identInfo
	data    interface{}
}

func (m *modelInfo) AddIndex(f *reflect.StructField, v *reflect.Value, indexType string, override bool) {
	fieldName := f.Name
	if _, ok := m.Indexes[fieldName]; !ok || override {
		m.Indexes[fieldName] = indexInfo{
			Type:  indexType,
			Field: f,
			Value: v,
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

func extract(s *reflect.Value, mi ...*modelInfo) (*modelInfo, error) {
	if s.Kind() == reflect.Ptr {
		e := s.Elem()
		s = &e
	}
	if s.Kind() != reflect.Struct {
		return nil, ErrBadType
	}

	typ := s.Type()

	var child bool

	var m *modelInfo
	if len(mi) > 0 {
		m = mi[0]
		child = true
	} else {
		m = &modelInfo{}
		m.Indexes = make(map[string]indexInfo)
		if !s.CanAddr() {
			return nil, ErrUnAddressable
		}
		m.data = s.Addr().Interface()
	}

	if m.Name == "" {
		m.Name = typ.Name()
	}

	numFields := s.NumField()
	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		value := s.Field(i)

		if field.PkgPath != "" {
			continue
		}

		err := extractField(&value, &field, m, child)
		if err != nil {
			return nil, err
		}
	}

	// ID field or tag detected
	if m.ID.Field != nil {
		zero := reflect.Zero(m.ID.Value.Type()).Interface()
		current := m.ID.Value.Interface()
		if reflect.DeepEqual(current, zero) {
			m.ID.IsZero = true
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

func extractField(value *reflect.Value, field *reflect.StructField, m *modelInfo, isChild bool) error {
	tag := field.Tag.Get("storm")
	if tag != "" {
		switch tag {
		case "id":
			m.ID.Field = field
			m.ID.Value = value
		case tagUniqueIdx, tagIdx:
			m.AddIndex(field, value, tag, !isChild)
		case tagInline:
			if value.Kind() == reflect.Ptr {
				e := value.Elem()
				value = &e
			}
			if value.Kind() == reflect.Struct {
				a := value.Addr()
				_, err := extract(&a, m)
				if err != nil {
					return err
				}
			}
		default:
			return ErrUnknownTag
		}
	}

	// the field is named ID and no ID field has been detected before
	if field.Name == "ID" && m.ID.Field == nil {
		m.ID.Field = field
		m.ID.Value = value
	}

	return nil
}

// Prefill the most requested informations
type identInfo struct {
	Field  *reflect.StructField
	Value  *reflect.Value
	IsZero bool
}

func (i *identInfo) Type() reflect.Type {
	return i.Value.Type()
}

func (i *identInfo) IsOfIntegerFamily() bool {
	return i.Value != nil && i.Value.Kind() >= reflect.Int && i.Value.Kind() <= reflect.Uint64
}

func getIndex(bucket *bolt.Bucket, idxKind string, fieldName string) (index.Index, error) {
	var idx index.Index
	var err error

	switch idxKind {
	case tagUniqueIdx:
		idx, err = index.NewUniqueIndex(bucket, []byte(indexPrefix+fieldName))
	case tagIdx:
		idx, err = index.NewListIndex(bucket, []byte(indexPrefix+fieldName))
	default:
		err = ErrIdxNotFound
	}

	return idx, err
}
