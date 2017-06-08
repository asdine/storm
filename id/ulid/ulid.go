// Package ulid provides an Universally Unique Lexicographically Sortable Identifier.
// Note that lexicographical order is only guaranteed in millisecond precision.
// See https://github.com/oklog/ulid
package ulid

import (
	"encoding/hex"
	"fmt"
	"sync"

	"time"

	"math/rand"

	"github.com/asdine/storm/id"
	"github.com/oklog/ulid"
)

// rand.Rand is not thread safe.
var randPool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	},
}

// New is the ULID ID provider.
var New id.New = func(start interface{}) id.Provider {
	return func(last []byte) (interface{}, error) {
		entropy := randPool.Get().(*rand.Rand)
		id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
		randPool.Put(entropy)
		return id, nil
	}
}
