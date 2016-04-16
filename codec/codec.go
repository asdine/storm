package codec

// EncodeDecoder represents a codec used to encode and decode entities.
type EncodeDecoder interface {
	Encode(v interface{}) ([]byte, error)
	Decode(b []byte, v interface{}) error
}
