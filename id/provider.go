// Package id contains sub-packages with different ID providers.
package id

// Provider is the func custom ID providers must implement.
// The last value will be provided by Storm.
// See https://golang.org/pkg/encoding/#BinaryUnmarshaler
// TODO(bep) this sounds wasteful when last is thrown away.
type Provider func(last []byte) (interface{}, error)

// New creates a new ID Provider.
type New func(start interface{}) Provider
