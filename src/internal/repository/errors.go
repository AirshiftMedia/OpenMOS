package repository

import "errors"

// Errors shared by all repository implementations. Implementations wrap them
// (fmt.Errorf("...: %w", ErrNotFound)), so callers test with errors.Is.
var (
	// ErrNotFound means the requested record, or the parent a new record
	// refers to, does not exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists means Create was called with a key that is in use.
	ErrAlreadyExists = errors.New("already exists")

	// ErrVersionConflict means an update was based on a stale version of a
	// running order.
	ErrVersionConflict = errors.New("version conflict")

	// ErrInvalid means the request itself is malformed: an empty or too long
	// ID, a value outside its MOS limits, or an ordering request that names
	// the same record twice.
	ErrInvalid = errors.New("invalid")
)
