package internal

import (
	"encoding/gob"
	"reflect"
	"testing"

	"github.com/asdine/storm/codec"
)

type testStruct struct {
	Name string
}

// RoundtripTester is a test helper to test a MarshalUnmarshaler
func RoundtripTester(t *testing.T, c codec.MarshalUnmarshaler, vals ...interface{}) {
	var val, to interface{}
	if len(vals) > 0 {
		if len(vals) != 2 {
			panic("Wrong number of vals, expected 2")
		}
		val = vals[0]
		to = vals[1]
	} else {
		val = &testStruct{Name: "test"}
		to = &testStruct{}
	}

	encoded, err := c.Marshal(val)
	if err != nil {
		t.Fatal("Encode error:", err)
	}
	err = c.Unmarshal(encoded, to)
	if err != nil {
		t.Fatal("Decode error:", err)
	}
	if !reflect.DeepEqual(val, to) {
		t.Fatalf("Roundtrip codec mismatch, expected\n%#v\ngot\n%#v", val, to)
	}
}

func init() {
	gob.Register(&testStruct{})
}
