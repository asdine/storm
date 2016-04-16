// Package codec contains sub-packages with different codecs that can be used
// to encode and decode entities in Storm.
package codec

import (
	"testing"

	"reflect"
)

type testStruct struct {
	Name string
}

// RountripTester is a test helper to test a EncodeDecoder
func RountripTester(t *testing.T, c EncodeDecoder, vals ...interface{}) {
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

	encoded, err := c.Encode(&val)
	if err != nil {
		t.Fatalf("Encode error:", err)
	}
	err = c.Decode(encoded, to)
	if err != nil {
		t.Fatal("Decode error:", err)
	}
	if !reflect.DeepEqual(val, to) {
		t.Fatalf("Roundtrip codec mismatch, expected\n%#v\ngot\n%#v", val, to)
	}
}
