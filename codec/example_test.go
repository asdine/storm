package codec_test

import (
	"fmt"

	"github.com/AndersonBargas/rainstorm/v5"
	"github.com/AndersonBargas/rainstorm/v5/codec/gob"
	"github.com/AndersonBargas/rainstorm/v5/codec/json"
	"github.com/AndersonBargas/rainstorm/v5/codec/msgpack"
	"github.com/AndersonBargas/rainstorm/v5/codec/protobuf"
	"github.com/AndersonBargas/rainstorm/v5/codec/sereal"
)

func Example() {
	// The examples below show how to set up all the codecs shipped with Rainstorm.
	// Proper error handling left out to make it simple.
	var gobDb, _ = rainstorm.Open("gob.db", rainstorm.Codec(gob.Codec))
	var jsonDb, _ = rainstorm.Open("json.db", rainstorm.Codec(json.Codec))
	var msgpackDb, _ = rainstorm.Open("msgpack.db", rainstorm.Codec(msgpack.Codec))
	var serealDb, _ = rainstorm.Open("sereal.db", rainstorm.Codec(sereal.Codec))
	var protobufDb, _ = rainstorm.Open("protobuf.db", rainstorm.Codec(protobuf.Codec))

	fmt.Printf("%T\n", gobDb.Codec())
	fmt.Printf("%T\n", jsonDb.Codec())
	fmt.Printf("%T\n", msgpackDb.Codec())
	fmt.Printf("%T\n", serealDb.Codec())
	fmt.Printf("%T\n", protobufDb.Codec())

	// Output:
	// *gob.gobCodec
	// *json.jsonCodec
	// *msgpack.msgpackCodec
	// *sereal.serealCodec
	// *protobuf.protobufCodec
}
