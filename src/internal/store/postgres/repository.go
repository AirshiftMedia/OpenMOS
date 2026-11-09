package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"airshift/openmos/internal/repository"
)

// DBTX is what the repositories need from a database handle. Both
// *pgxpool.Pool and pgx.Tx satisfy it; on a pgx.Tx, Begin starts a
// savepoint, so the repositories also work inside a caller's transaction.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Repository implements repository.Repository on the openmos schema. The
// queries are written by hand; the behaviour is the contract in package
// repository, checked by the repotest suite.
type Repository struct{ db DBTX }

var _ repository.Repository = (*Repository)(nil)

// New returns a repository using db, normally a *pgxpool.Pool from NewPool.
// The schema must be migrated.
func New(db DBTX) *Repository { return &Repository{db: db} }

func (r *Repository) RunningOrders() repository.RunningOrderRepository { return runningOrders{r.db} }
func (r *Repository) Stories() repository.StoryRepository              { return stories{r.db} }
func (r *Repository) Items() repository.ItemRepository                 { return items{r.db} }
func (r *Repository) Objects() repository.ObjectRepository             { return objects{r.db} }

// InTx runs fn in a database transaction, committed if fn returns nil.
// On a repository already bound to a transaction it uses a savepoint.
// Isolation is READ COMMITTED: fn sees rows others commit meanwhile, while
// the version check on running orders and the parent locks guard against
// lost updates.
func (r *Repository) InTx(ctx context.Context, fn func(repository.Repository) error) error {
	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error { return fn(New(tx)) })
}

// dbError maps a database error to the repository errors, keeping the
// original in the chain. what names the record, as in "story ro-1/s1".
func dbError(err error, what string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", what, repository.ErrNotFound)
	}
	var pe *pgconn.PgError
	if errors.As(err, &pe) {
		var sentinel error
		switch pe.Code {
		case "23505": // unique_violation
			sentinel = repository.ErrAlreadyExists
		case "23503": // foreign_key_violation: the parent is gone
			sentinel = repository.ErrNotFound
		case "23514", // check_violation: a value outside its MOS limits
			"22003", // numeric_value_out_of_range
			"22021", // character_not_in_repertoire, e.g. a NUL byte
			"22P05": // untranslatable_character
			sentinel = repository.ErrInvalid
		}
		if sentinel != nil {
			return fmt.Errorf("%s: %w: %w", what, sentinel, err)
		}
	}
	return fmt.Errorf("%s: %w", what, err)
}

// metadata returns m, or an empty map for nil: the column is NOT NULL and
// a nil map would encode as JSON null.
func metadata(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}

// siblings describes the children of one parent: the stories of a running
// order or the items of a story. Its SQL takes the parent key as the first
// len(key) parameters.
type siblings struct {
	table string // e.g. openmos.stories
	idCol string // e.g. story_id
	where string // selects the children of the parent, e.g. ro_id = $1
	lock  string // locks the parent row; no row means no parent
}

// lockParent locks the parent row against concurrent changes to its
// children's positions, and reports ErrNotFound if it does not exist. All
// operations that assign positions take this lock first, so appends,
// moves and replaces of one parent run one at a time.
func (s siblings) lockParent(ctx context.Context, tx pgx.Tx, what string, key ...any) error {
	var one int
	if err := tx.QueryRow(ctx, s.lock, key...).Scan(&one); err != nil {
		return dbError(err, what)
	}
	return nil
}

// nextPosition is a subquery giving the position after the last child.
func (s siblings) nextPosition() string {
	return "(SELECT COALESCE(MAX(position) + 1, 0) FROM " + s.table + " WHERE " + s.where + ")"
}

// reorder locks the parent, passes the current child IDs in order to
// compute and stores the order it returns. compute returns an order of the
// same IDs (see repository.MoveIDs and SwapIDs).
func (s siblings) reorder(ctx context.Context, db DBTX, what string, key []any, compute func([]string) ([]string, error)) error {
	return pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		if err := s.lockParent(ctx, tx, what, key...); err != nil {
			return err
		}
		rows, err := tx.Query(ctx,
			"SELECT "+s.idCol+" FROM "+s.table+" WHERE "+s.where+" ORDER BY position, "+s.idCol, key...)
		if err != nil {
			return dbError(err, what)
		}
		ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
		if err != nil {
			return dbError(err, what)
		}
		order, err := compute(ids)
		if err != nil {
			return fmt.Errorf("%s: %w", what, err)
		}
		// One statement, so the deferrable position constraint is checked
		// only after every child has its new position.
		n := len(key) + 1
		_, err = tx.Exec(ctx, fmt.Sprintf(
			"UPDATE %s SET position = u.ord - 1 FROM unnest($%d::text[]) WITH ORDINALITY AS u(id, ord) WHERE %s AND %s = u.id",
			s.table, n, s.where, s.idCol), append(key, order)...)
		if err != nil {
			return dbError(err, what)
		}
		return nil
	})
}
