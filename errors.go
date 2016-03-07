package storm

import "errors"

// Errors
var (
	ErrNoID = errors.New("missing struct tag id or ID field")
)
