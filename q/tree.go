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

// A ValueMatcher is used to test against a reflect.Value.
type ValueMatcher interface {
	// MatchValue tests if the given reflect.Value matches.
	// It is useful when the reflect.Value of an object already exists.
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

type trueMatcher struct{}

func (*trueMatcher) Match(i interface{}) bool {
	return true
}

func (*trueMatcher) MatchValue(v *reflect.Value) bool {
	return true
}

type or struct {
	children []Matcher
}

func (c *or) Match(i interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(i))
	return c.MatchValue(&v)
}

func (c *or) MatchValue(v *reflect.Value) bool {
	for _, matcher := range c.children {
		if vm, ok := matcher.(ValueMatcher); ok {
			if vm.MatchValue(v) {
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
	return c.MatchValue(&v)
}

func (c *and) MatchValue(v *reflect.Value) bool {
	for _, matcher := range c.children {
		if vm, ok := matcher.(ValueMatcher); ok {
			if !vm.MatchValue(v) {
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

type strictEq struct {
	field string
	value interface{}
}

func (s *strictEq) Match(i interface{}) bool {
	v := reflect.Indirect(reflect.ValueOf(i))
	return s.MatchValue(&v)
}

func (s *strictEq) MatchValue(v *reflect.Value) bool {
	field := v.FieldByName(s.field)
	return reflect.DeepEqual(field.Interface(), s.value)
}

// Eq matcher, checks if the given field is equal to the given value
func Eq(field string, v interface{}) Matcher { return &cmp{field: field, value: v, token: token.EQL} }

// StrictEq matcher, checks if the given field is deeply equal to the given value
func StrictEq(field string, v interface{}) Matcher { return &strictEq{field: field, value: v} }

// Gt matcher, checks if the given field is greater than the given value
func Gt(field string, v interface{}) Matcher { return &cmp{field: field, value: v, token: token.GTR} }

// Gte matcher, checks if the given field is greater than or equal to the given value
func Gte(field string, v interface{}) Matcher { return &cmp{field: field, value: v, token: token.GEQ} }

// Lt matcher, checks if the given field is lesser than the given value
func Lt(field string, v interface{}) Matcher { return &cmp{field: field, value: v, token: token.LSS} }

// Lte matcher, checks if the given field is lesser than or equal to the given value
func Lte(field string, v interface{}) Matcher { return &cmp{field: field, value: v, token: token.LEQ} }

// True matcher, always returns true
func True() Matcher { return &trueMatcher{} }

// Or matcher, checks if at least one of the given matchers matches the record
func Or(matchers ...Matcher) Matcher { return &or{children: matchers} }

// And matcher, checks if all of the given matchers matches the record
func And(matchers ...Matcher) Matcher { return &and{children: matchers} }
