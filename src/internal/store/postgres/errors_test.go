package postgres

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"airshift/openmos/internal/repository"
)

// dbError is also reached by paths the contract suite cannot provoke, such
// as a foreign key violation, which the parent locks normally prevent.
func TestDBError(t *testing.T) {
	other := errors.New("connection reset")
	for _, tc := range []struct {
		err  error
		want error
	}{
		{pgx.ErrNoRows, repository.ErrNotFound},
		{fmt.Errorf("scan: %w", pgx.ErrNoRows), repository.ErrNotFound},
		{&pgconn.PgError{Code: "23505"}, repository.ErrAlreadyExists},
		{&pgconn.PgError{Code: "23503"}, repository.ErrNotFound},
		{&pgconn.PgError{Code: "23514"}, repository.ErrInvalid},
		{&pgconn.PgError{Code: "22003"}, repository.ErrInvalid},
		{&pgconn.PgError{Code: "22021"}, repository.ErrInvalid},
		{&pgconn.PgError{Code: "22P05"}, repository.ErrInvalid},
		{&pgconn.PgError{Code: "40001"}, nil},
		{other, nil},
	} {
		got := dbError(tc.err, "story ro/s")
		// No rows is just "not found"; other causes stay in the chain
		if !errors.Is(tc.err, pgx.ErrNoRows) && !errors.Is(got, tc.err) {
			t.Errorf("dbError(%v) = %v, lost the original error", tc.err, got)
		}
		for _, sentinel := range []error{repository.ErrNotFound, repository.ErrAlreadyExists,
			repository.ErrInvalid, repository.ErrVersionConflict} {
			if errors.Is(got, sentinel) != (sentinel == tc.want) {
				t.Errorf("dbError(%v) = %v; errors.Is(%v) = %v", tc.err, got, sentinel, !(sentinel == tc.want))
			}
		}
	}
}
