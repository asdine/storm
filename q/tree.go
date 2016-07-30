package q

import (
	"go/token"
	"reflect"
)

// A Criteria is used to test against a record to see if it matches.
// It can be a combination of multiple other criterias.
type Criteria interface {
	// Match is used to test the criteria against a structure.
	Match(interface{}) bool

	// MatchValue is useful when the reflect.Value of the structure already exists.
	// It is used internally to avoid recreating a reflect.Value when executing a tree of Criteria
	MatchValue(*reflect.Value) bool
}

type cmp struct {
	field string
	value interface{}
	token token.Token
}

func (c *cmp) Match(i interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(i))
	return c.MatchValue(&v)
}

func (c *cmp) MatchValue(v *reflect.Value) bool {
	field := v.FieldByName(c.field)
	return compare(field.Interface(), c.value, c.token)
}

type or struct {
	children []Criteria
}

func (c *or) Match(i interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(i))
	return c.MatchValue(&v)
}

func (c *or) MatchValue(v *reflect.Value) bool {
	for _, criteria := range c.children {
		if criteria.MatchValue(v) {
			return true
		}
	}

	return false
}

type and struct {
	children []Criteria
}

func (c *and) Match(i interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(i))
	return c.MatchValue(&v)
}

func (c *and) MatchValue(v *reflect.Value) bool {
	for _, criteria := range c.children {
		if !criteria.MatchValue(v) {
			return false
		}
	}

	return true
}

// Eq criteria, checks if the given field is equal to the given value
func Eq(field string, v interface{}) Criteria { return &cmp{field: field, value: v, token: token.EQL} }

// Gt criteria, checks if the given field is greater than the given value
func Gt(field string, v interface{}) Criteria { return &cmp{field: field, value: v, token: token.GTR} }

// Gte criteria, checks if the given field is greater than or equal to the given value
func Gte(field string, v interface{}) Criteria { return &cmp{field: field, value: v, token: token.GEQ} }

// Lt criteria, checks if the given field is lesser than the given value
func Lt(field string, v interface{}) Criteria { return &cmp{field: field, value: v, token: token.LSS} }

// Lte criteria, checks if the given field is lesser than or equal to the given value
func Lte(field string, v interface{}) Criteria { return &cmp{field: field, value: v, token: token.LEQ} }

// Or criteria, checks if at least one of the given criterias matches the record
func Or(criterias ...Criteria) Criteria { return &or{children: criterias} }

// And criteria, checks if all of the given criterias matches the record
func And(criterias ...Criteria) Criteria { return &and{children: criterias} }
