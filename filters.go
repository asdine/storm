package storm

import (
	"fmt"
	"go/constant"
	"go/token"
	"reflect"
	"strconv"

	"github.com/asdine/storm/codec"
	"github.com/boltdb/bolt"
)

type filter interface {
	exec(i interface{}) bool
}

type operator interface {
	filter
	setLeftFilter(l filter)
	setRightFilter(l filter)
}

type filterAnd struct {
	Left  filter
	Right filter
}

func (f *filterAnd) exec(i interface{}) bool {
	return f.Left.exec(i) && f.Right.exec(i)
}

func (f *filterAnd) setLeftNode(l filter) {
	f.Left = l
}

func (f *filterAnd) setRightNode(r filter) {
	f.Right = r
}

type filterOr struct {
	Left  filter
	Right filter
}

func (f *filterOr) exec(i interface{}) bool {
	return f.Left.exec(i) || f.Right.exec(i)
}

func (f *filterOr) setLeftNode(l filter) {
	f.Left = l
}

func (f *filterOr) setRightNode(r filter) {
	f.Right = r
}

type filterEq struct {
	field string
	value interface{}
}

func (f *filterEq) exec(i interface{}) bool {
	v := reflect.ValueOf(i)
	field := v.FieldByName(f.field)
	return compare(field, f.value, token.EQL)
}

type filterStrictEq struct {
	field string
	value interface{}
}

func (f *filterStrictEq) exec(i interface{}) bool {
	v := reflect.ValueOf(i)
	field := v.FieldByName(f.field)
	return reflect.DeepEqual(field.Interface(), f.value)
}

type filterIn struct {
	field string
	value []interface{}
}

func (f *filterIn) exec(i interface{}) bool {
	v := reflect.ValueOf(i)
	field := v.FieldByName(f.field)

	for i := range f.value {
		if compare(field.Interface(), f.value[i], token.EQL) {
			return true
		}
	}
	return false
}

type filterNotIn struct {
	field string
	value []interface{}
}

func (f *filterNotIn) exec(i interface{}) bool {
	v := reflect.ValueOf(i)
	field := v.FieldByName(f.field)

	for i := range f.value {
		if compare(field.Interface(), f.value[i], token.EQL) {
			return false
		}
	}
	return true
}

type filterCmp struct {
	field string
	value interface{}
	tok   token.Token
}

func (f *filterCmp) exec(i interface{}) bool {
	v := reflect.ValueOf(i)
	field := v.FieldByName(f.field)
	return compare(field.Interface(), f.value, f.tok)
}

func compare(a, b interface{}, tok token.Token) bool {
	vala := reflect.ValueOf(a)
	valb := reflect.ValueOf(b)

	ak := vala.Kind()
	bk := valb.Kind()
	switch {
	case ak >= reflect.Int && ak <= reflect.Int64:
		if bk >= reflect.Int && bk <= reflect.Int64 {
			return constant.Compare(constant.MakeInt64(vala.Int()), tok, constant.MakeInt64(valb.Int()))
		}

		if bk == reflect.Float32 || bk == reflect.Float64 {
			return constant.Compare(constant.MakeFloat64(float64(vala.Int())), tok, constant.MakeFloat64(valb.Float()))
		}

		if bk == reflect.String {
			bla, err := strconv.ParseFloat(valb.String(), 64)
			if err != nil {
				return false
			}

			return constant.Compare(constant.MakeFloat64(float64(vala.Int())), tok, constant.MakeFloat64(bla))
		}
	case ak == reflect.Float32 || ak == reflect.Float64:
		if bk == reflect.Float32 || bk == reflect.Float64 {
			return constant.Compare(constant.MakeFloat64(vala.Float()), tok, constant.MakeFloat64(valb.Float()))
		}

		if bk >= reflect.Int && bk <= reflect.Int64 {
			return constant.Compare(constant.MakeFloat64(vala.Float()), tok, constant.MakeFloat64(float64(valb.Int())))
		}

		if bk == reflect.String {
			bla, err := strconv.ParseFloat(valb.String(), 64)
			if err != nil {
				return false
			}

			return constant.Compare(constant.MakeFloat64(vala.Float()), tok, constant.MakeFloat64(bla))
		}
	case ak == reflect.String:
		if bk == reflect.String {
			return constant.Compare(constant.MakeString(vala.String()), tok, constant.MakeString(valb.String()))
		}
	case tok == token.EQL:
		return reflect.DeepEqual(a, b)
	}

	return false
}

func selector(b *bolt.Bucket, qtree filter, codec codec.EncodeDecoder, model interface{}) error {
	value := reflect.ValueOf(model)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	c := b.Cursor()

	for k, v := c.First(); k != nil; k, v = c.Next() {
		// skip buckets
		if v == nil {
			continue
		}

		err := codec.Decode(v, value.Addr().Interface())
		if err != nil {
			return err
		}

		if qtree.exec(value.Interface()) {
			fmt.Println("Found field", value.Interface())
		}
	}

	return nil
}
