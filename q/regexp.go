package q

import (
	"fmt"
	"reflect"
	"regexp"
	"sync"
)

// Re creates a regexp matcher. It checks if the given field matches the given regexp.
// Note that this only supports fields of type string or []byte.
func Re(field string, re string) Matcher {
	regexpCache.RLock()
	if r, ok := regexpCache.m[re]; ok {
		regexpCache.RUnlock()
		return &regexpMatcher{field: field, r: r}
	}
	regexpCache.RUnlock()

	regexpCache.Lock()
	r, err := regexp.Compile(re)
	if err == nil {
		regexpCache.m[re] = r
	}
	regexpCache.Unlock()

	return &regexpMatcher{field: field, r: r, err: err}
}

var regexpCache = struct {
	sync.RWMutex
	m map[string]*regexp.Regexp
}{m: make(map[string]*regexp.Regexp)}

type regexpMatcher struct {
	field string
	r     *regexp.Regexp
	err   error
}

func (r *regexpMatcher) Match(i interface{}) (bool, error) {
	v := reflect.Indirect(reflect.ValueOf(i))
	return r.MatchValue(&v)
}

func (r *regexpMatcher) MatchValue(v *reflect.Value) (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	field := v.FieldByName(r.field)

	switch fieldValue := field.Interface().(type) {
	case string:
		return r.r.MatchString(fieldValue), nil
	case []byte:
		return r.r.Match(fieldValue), nil
	default:
		return false, fmt.Errorf("Only string and []byte supported for regexp matcher, got %T", fieldValue)
	}
}
