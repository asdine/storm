package storm

import (
	"fmt"

	"github.com/fatih/structs"
)

type tags struct {
	Name    string
	ID      interface{}
	IDField *structs.Field
	ZeroID  bool
	Uniques []*structs.Field
	Indexes []*structs.Field
}

func extractTags(data interface{}, tg ...*tags) (*tags, error) {
	s := structs.New(data)
	fields := s.Fields()

	var t *tags
	if len(tg) > 0 {
		t = tg[0]
	} else {
		t = &tags{}
	}

	if t.Name == "" {
		t.Name = s.Name()
	}

	for _, f := range fields {
		if !f.IsExported() {
			continue
		}

		tag := f.Tag("storm")
		if tag != "" {
			switch tag {
			case "id":
				if f.IsZero() {
					t.ZeroID = true
				}
				t.ID = f.Value()
				t.IDField = f
			case "unique":
				t.Uniques = append(t.Uniques, f)
			case "index":
				t.Indexes = append(t.Indexes, f)
			case "inline":
				if structs.IsStruct(f.Value()) {
					_, err := extractTags(f.Value(), t)
					if err != nil {
						return nil, err
					}
				}
			default:
				return nil, fmt.Errorf("unknown tag %s", tag)
			}
		}

		if f.Name() == "ID" && t.ID == nil {
			t.ID = f.Value()
			t.IDField = f
		}
	}

	if t.ID != nil {
		t.Uniques = append(t.Uniques, t.IDField)
	}

	return t, nil
}

func indexKind(index string, t *tags) string {
	for i := range t.Indexes {
		if t.Indexes[i].Name() == index {
			return "list"
		}
	}

	for i := range t.Uniques {
		if t.Uniques[i].Name() == index {
			return "unique"
		}
	}

	return ""
}
