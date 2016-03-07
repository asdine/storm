package storm

import "errors"

// Errors
var (
	ErrNoID          = errors.New("missing struct tag id or ID field")
	ErrZeroID        = errors.New("id field must not be a zero value")
	ErrBadType       = errors.New("provided data must be a struct or a pointer to struct")
	ErrAlreadyExists = errors.New("already exists")
	ErrNilParam      = errors.New("param must not be nil")
)
