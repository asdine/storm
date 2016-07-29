package q

import (
	"go/token"
	"reflect"
)

type Criteria interface {
	Exec(interface{}) bool
}

type eq struct {
	field string
	value interface{}
}

func (c *eq) Exec(i interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(i))

	field := v.FieldByName(c.field)
	return compare(field.Interface(), c.value, token.EQL)
}

type or struct {
	children []Criteria
}

func (c *or) Exec(i interface{}) bool {
	for _, criteria := range c.children {
		if criteria.Exec(i) {
			return true
		}
	}

	return false
}

type and struct {
	children []Criteria
}

func (c *and) Exec(i interface{}) bool {
	for _, criteria := range c.children {
		if !criteria.Exec(i) {
			return false
		}
	}

	return true
}

func Eq(field string, v interface{}) Criteria { return &eq{field: field, value: v} }

func Or(criterias ...Criteria) Criteria  { return &or{children: criterias} }
func And(criterias ...Criteria) Criteria { return &and{children: criterias} }
