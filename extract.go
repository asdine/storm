package storm

import (
	"reflect"
	"strings"

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

type fieldConfig struct {
	Name      string
	Type      string
	IsZero    bool
	IsID      bool
	Increment bool
	IsInteger bool
	Value     *reflect.Value
}

// structConfig is a structure gathering all the relevant informations about a model
type structConfig struct {
	Name   string
	Fields map[string]*fieldConfig
	ID     *fieldConfig
}

// helper
func (m *structConfig) AllByType(indexType string) []*fieldConfig {
	var idx []*fieldConfig
	for k := range m.Fields {
		if m.Fields[k].Type == indexType {
			idx = append(idx, m.Fields[k])
		}
	}

	return idx
}

func extract(s *reflect.Value, mi ...*structConfig) (*structConfig, error) {
	if s.Kind() == reflect.Ptr {
		e := s.Elem()
		s = &e
	}
	if s.Kind() != reflect.Struct {
		return nil, ErrBadType
	}

	typ := s.Type()

	var child bool

	var m *structConfig
	if len(mi) > 0 {
		m = mi[0]
		child = true
	} else {
		m = &structConfig{}
		m.Fields = make(map[string]*fieldConfig)
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
	if m.ID != nil {
		zero := reflect.Zero(m.ID.Value.Type()).Interface()
		current := m.ID.Value.Interface()
		if reflect.DeepEqual(current, zero) {
			m.ID.IsZero = true
		}
	}

	if child {
		return m, nil
	}

	if m.ID == nil {
		return nil, ErrNoID
	}

	if m.Name == "" {
		return nil, ErrNoName
	}

	return m, nil
}

func extractField(value *reflect.Value, field *reflect.StructField, m *structConfig, isChild bool) error {
	var f *fieldConfig

	tag := field.Tag.Get("storm")
	if tag != "" {
		f = &fieldConfig{
			Name:      field.Name,
			IsZero:    isZero(value),
			IsInteger: isInteger(value),
			Value:     value,
		}

		tags := strings.Split(tag, ",")

		for _, tag := range tags {
			switch tag {
			case "id":
				f.IsID = true
			case tagUniqueIdx, tagIdx:
				f.Type = tag
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
				// we don't need to save this field
				return nil
			default:
				return ErrUnknownTag
			}
		}

		if f.Type != "" {
			if _, ok := m.Fields[f.Name]; !ok || !isChild {
				m.Fields[f.Name] = f
			}
		}
	}

	if m.ID == nil && f != nil && f.IsID {
		m.ID = f
	}

	// the field is named ID and no ID field has been detected before
	if m.ID == nil && field.Name == "ID" {
		if f == nil {
			f = &fieldConfig{
				Name:      field.Name,
				IsZero:    isZero(value),
				IsInteger: isInteger(value),
				Value:     value,
			}
		}
		m.ID = f
	}

	return nil
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

func isZero(v *reflect.Value) bool {
	zero := reflect.Zero(v.Type()).Interface()
	current := v.Interface()
	return reflect.DeepEqual(current, zero)
}

func isInteger(v *reflect.Value) bool {
	kind := v.Kind()
	return v != nil && kind >= reflect.Int && kind <= reflect.Uint64
}
