// Package codec contains sub-packages with different codecs that can be used
// to encode and decode entities in Storm.
package codec

// EncodeDecoder represents a codec used to encode and decode entities.
type EncodeDecoder interface {
	Encode(v interface{}) ([]byte, error)
	Decode(b []byte, v interface{}) error
}
