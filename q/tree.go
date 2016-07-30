package q

import (
	"go/token"
	"reflect"
)

// A Matcher is used to test against a record to see if it matches.
type Matcher interface {
	// Match is used to test the criteria against a structure.
	Match(interface{}) bool
}

type valueMatcher interface {
	// matchValue tests if the given reflect.Value matches.
	// It is useful when the reflect.Value of an object already exists.
	matchValue(*reflect.Value) bool
}

type cmp struct {
	field string
	value interface{}
	token token.Token
}

func (c *cmp) Match(i interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(i))
	return c.matchValue(&v)
}

func (c *cmp) matchValue(v *reflect.Value) bool {
	field := v.FieldByName(c.field)
	return compare(field.Interface(), c.value, c.token)
}

type or struct {
	children []Matcher
}

func (c *or) Match(i interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(i))
	return c.matchValue(&v)
}

func (c *or) matchValue(v *reflect.Value) bool {
	for _, matcher := range c.children {
		if vm, ok := matcher.(valueMatcher); ok {
			if vm.matchValue(v) {
				return true
			}
			continue
		}

		if matcher.Match(v.Interface()) {
			return true
		}
	}

	return false
}

type and struct {
	children []Matcher
}

func (c *and) Match(i interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(i))
	return c.matchValue(&v)
}

func (c *and) matchValue(v *reflect.Value) bool {
	for _, matcher := range c.children {
		if vm, ok := matcher.(valueMatcher); ok {
			if !vm.matchValue(v) {
				return false
			}
			continue
		}

		if !matcher.Match(v.Interface()) {
			return false
		}
	}

	return true
}

// Eq criteria, checks if the given field is equal to the given value
func Eq(field string, v interface{}) Matcher { return &cmp{field: field, value: v, token: token.EQL} }

// Gt criteria, checks if the given field is greater than the given value
func Gt(field string, v interface{}) Matcher { return &cmp{field: field, value: v, token: token.GTR} }

// Gte criteria, checks if the given field is greater than or equal to the given value
func Gte(field string, v interface{}) Matcher { return &cmp{field: field, value: v, token: token.GEQ} }

// Lt criteria, checks if the given field is lesser than the given value
func Lt(field string, v interface{}) Matcher { return &cmp{field: field, value: v, token: token.LSS} }

// Lte criteria, checks if the given field is lesser than or equal to the given value
func Lte(field string, v interface{}) Matcher { return &cmp{field: field, value: v, token: token.LEQ} }

// Or criteria, checks if at least one of the given criterias matches the record
func Or(criterias ...Matcher) Matcher { return &or{children: criterias} }

// And criteria, checks if all of the given criterias matches the record
func And(criterias ...Matcher) Matcher { return &and{children: criterias} }
