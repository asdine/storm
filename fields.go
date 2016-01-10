package storm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/structs"
)

type tags struct {
	Name    string
	ID      interface{}
	IDField interface{}
	Uniques []*structs.Field
	Indexes []*structs.Field
}

func extractTags(data interface{}) (*tags, error) {
	s := structs.New(data)
	fields := s.Fields()

	var t tags
	t.Name = strings.ToLower(s.Name())

	for _, f := range fields {
		if !f.IsExported() {
			continue
		}

		tag := f.Tag("storm")
		if tag != "" {
			switch tag {
			case "id":
				if f.IsZero() {
					return nil, errors.New("id field must not be a zero value")
				}
				t.ID = f.Value()
			case "unique":
				t.Uniques = append(t.Uniques, f)
			case "index":
				t.Indexes = append(t.Indexes, f)
			default:
				return nil, fmt.Errorf("unknown tag %s", tag)
			}
		}

		if f.Name() == "ID" {
			t.IDField = f.Value()
		}
	}

	return &t, nil
}
