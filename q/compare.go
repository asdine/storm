package q

import (
	"go/constant"
	"go/token"
	"reflect"
	"strconv"
)

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
