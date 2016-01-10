package storm

import (
	"errors"

	"github.com/fatih/structs"
)

type tags struct {
	Name    string
	ID      interface{}
	IDField interface{}
}

func extractTags(data interface{}) (*tags, error) {
	s := structs.New(data)
	fields := s.Fields()

	var t tags
	t.Name = s.Name()

	for _, f := range fields {
		if !f.IsExported() {
			continue
		}

		tag := f.Tag("storm")
		if tag == "id" {
			if f.IsZero() {
				return nil, errors.New("id field must not be a zero value")
			}
			t.ID = f.Value()
		}

		if f.Name() == "ID" {
			t.IDField = f.Value()
		}
	}

	return &t, nil
}
