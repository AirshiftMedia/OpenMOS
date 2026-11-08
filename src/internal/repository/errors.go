package repository

import "errors"

// Errors shared by all repository implementations. Implementations wrap them
// (fmt.Errorf("...: %w", ErrNotFound)), so callers test with errors.Is.
var (
	// ErrNotFound means the requested record, or the parent a new record
	// refers to, does not exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists means Create was called with an ID that is in use.
	ErrAlreadyExists = errors.New("already exists")

	// ErrVersionConflict means an update was based on a stale version of a
	// running order.
	ErrVersionConflict = errors.New("version conflict")
)
