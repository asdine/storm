package storm

import "github.com/asdine/storm/codec"

// toBytes turns an interface into a slice of bytes
func toBytes(key interface{}, encoder codec.EncodeDecoder) ([]byte, error) {
	if key == nil {
		return nil, nil
	}
	if k, ok := key.([]byte); ok {
		return k, nil
	}
	if k, ok := key.(string); ok {
		return []byte(k), nil
	}

	return encoder.Encode(key)
}
