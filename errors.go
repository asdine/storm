package storm

import "errors"

// Errors
var (
	ErrNoID            = errors.New("missing struct tag id or ID field")
	ErrZeroID          = errors.New("id field must not be a zero value")
	ErrBadType         = errors.New("provided data must be a struct or a pointer to struct")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNilParam        = errors.New("param must not be nil")
	ErrBadIndexType    = errors.New("bad index type")
	ErrSlicePtrNeeded  = errors.New("provided target must be a pointer to slice")
	ErrStructPtrNeeded = errors.New("provided target must be a pointer to struct")
	ErrNoName          = errors.New("provided target must have a name")
	ErrIndexNotFound   = errors.New("index not found")
	ErrNotFound        = errors.New("not found")
)
