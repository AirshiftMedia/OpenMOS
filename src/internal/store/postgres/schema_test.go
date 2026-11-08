package postgres

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL error codes the schema is expected to raise.
const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
	checkViolation      = "23514"
)

// wantPgError fails the test unless stmt fails with the given SQLSTATE code.
func wantPgError(t *testing.T, db *sql.DB, code, stmt string) {
	t.Helper()
	_, err := db.ExecContext(t.Context(), stmt)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != code {
		t.Errorf("%s:\n err = %v, want SQLSTATE %s", stmt, err, code)
	}
}

// seed creates ro-1 with stories s1 (items i1, i2) and s2 (item i3).
func seed(t *testing.T, db *sql.DB) {
	t.Helper()
	exec(t, db,
		`INSERT INTO openmos.running_orders (ro_id) VALUES ('ro-1')`,
		`INSERT INTO openmos.stories (ro_id, story_id, position) VALUES ('ro-1', 's1', 0), ('ro-1', 's2', 1)`,
		`INSERT INTO openmos.items (ro_id, story_id, item_id, position)
			VALUES ('ro-1', 's1', 'i1', 0), ('ro-1', 's1', 'i2', 1), ('ro-1', 's2', 'i3', 0)`,
	)
}

func TestCompositeKeys(t *testing.T) {
	db := migrated(t)
	seed(t, db)

	// The same story and item IDs may appear under another parent
	exec(t, db,
		`INSERT INTO openmos.running_orders (ro_id) VALUES ('ro-2')`,
		`INSERT INTO openmos.stories (ro_id, story_id, position) VALUES ('ro-2', 's1', 0)`,
		`INSERT INTO openmos.items (ro_id, story_id, item_id, position) VALUES ('ro-2', 's1', 'i1', 0)`,
		`INSERT INTO openmos.mos_objects (mos_id, obj_id) VALUES ('a.mos', 'o1'), ('b.mos', 'o1')`,
	)
	// but not twice under the same one
	wantPgError(t, db, uniqueViolation, `INSERT INTO openmos.stories (ro_id, story_id, position) VALUES ('ro-1', 's1', 5)`)
	wantPgError(t, db, uniqueViolation, `INSERT INTO openmos.items (ro_id, story_id, item_id, position) VALUES ('ro-1', 's1', 'i1', 5)`)
	wantPgError(t, db, uniqueViolation, `INSERT INTO openmos.mos_objects (mos_id, obj_id) VALUES ('a.mos', 'o1')`)

	// and a child needs its parent
	wantPgError(t, db, foreignKeyViolation, `INSERT INTO openmos.stories (ro_id, story_id, position) VALUES ('missing', 's1', 0)`)
	wantPgError(t, db, foreignKeyViolation, `INSERT INTO openmos.items (ro_id, story_id, item_id, position) VALUES ('ro-1', 'missing', 'i1', 0)`)
}

func TestPositions(t *testing.T) {
	db := migrated(t)
	seed(t, db)

	wantPgError(t, db, uniqueViolation, `INSERT INTO openmos.stories (ro_id, story_id, position) VALUES ('ro-1', 's3', 1)`)
	wantPgError(t, db, uniqueViolation, `INSERT INTO openmos.items (ro_id, story_id, item_id, position) VALUES ('ro-1', 's1', 'i9', 0)`)

	// Swapping two stories one row at a time fails while the constraint is
	// checked per statement
	tx, err := db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.ExecContext(t.Context(), `UPDATE openmos.stories SET position = 1 WHERE ro_id = 'ro-1' AND story_id = 's1'`)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != uniqueViolation {
		t.Errorf("immediate swap step: err = %v, want SQLSTATE %s", err, uniqueViolation)
	}
	tx.Rollback()

	// but succeeds once deferred to commit
	tx, err = db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{
		`SET CONSTRAINTS openmos.stories_position_key DEFERRED`,
		`UPDATE openmos.stories SET position = 1 WHERE ro_id = 'ro-1' AND story_id = 's1'`,
		`UPDATE openmos.stories SET position = 0 WHERE ro_id = 'ro-1' AND story_id = 's2'`,
	} {
		if _, err := tx.ExecContext(t.Context(), s); err != nil {
			tx.Rollback()
			t.Fatalf("%s: %v", s, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit deferred swap: %v", err)
	}
	var first string
	if err := db.QueryRowContext(t.Context(),
		`SELECT story_id FROM openmos.stories WHERE ro_id = 'ro-1' ORDER BY position LIMIT 1`).Scan(&first); err != nil {
		t.Fatal(err)
	}
	if first != "s2" {
		t.Errorf("first story after swap = %s, want s2", first)
	}
}

func TestCascadeDeletes(t *testing.T) {
	db := migrated(t)
	seed(t, db)

	exec(t, db, `DELETE FROM openmos.stories WHERE ro_id = 'ro-1' AND story_id = 's1'`)
	if n := count(t, db, `SELECT count(*) FROM openmos.items WHERE story_id = 's1'`); n != 0 {
		t.Errorf("%d items left after deleting their story", n)
	}
	if n := count(t, db, `SELECT count(*) FROM openmos.items`); n != 1 {
		t.Errorf("%d items left, want 1 (i3)", n)
	}

	exec(t, db, `DELETE FROM openmos.running_orders WHERE ro_id = 'ro-1'`)
	if n := count(t, db, `SELECT count(*) FROM openmos.stories`) + count(t, db, `SELECT count(*) FROM openmos.items`); n != 0 {
		t.Errorf("%d stories and items left after deleting their running order", n)
	}
}

func TestMOSLimits(t *testing.T) {
	db := migrated(t)
	seed(t, db)
	long := strings.Repeat("x", 129)

	for _, stmt := range []string{
		`INSERT INTO openmos.running_orders (ro_id) VALUES ('')`,
		`INSERT INTO openmos.running_orders (ro_id) VALUES ('` + long + `')`,
		`INSERT INTO openmos.running_orders (ro_id, ed_dur) VALUES ('ro-neg', interval '-1 second')`,
		`INSERT INTO openmos.stories (ro_id, story_id, position) VALUES ('ro-1', 's9', -1)`,
		`UPDATE openmos.items SET ed_dur = -1 WHERE item_id = 'i1'`,
		`UPDATE openmos.items SET ed_dur = 4294967296 WHERE item_id = 'i1'`,
		`UPDATE openmos.items SET time_base = 0 WHERE item_id = 'i1'`,
		`INSERT INTO openmos.mos_objects (mos_id, obj_id, obj_type) VALUES ('a.mos', 'o1', 'FILM')`,
		`INSERT INTO openmos.mos_objects (mos_id, obj_id, air) VALUES ('a.mos', 'o1', 'MAYBE')`,
		`INSERT INTO openmos.mos_objects (mos_id, obj_id, revision) VALUES ('a.mos', 'o1', 1000)`,
	} {
		wantPgError(t, db, checkViolation, stmt)
	}

	// Values at the limits are accepted, and an NTSC time base is exact
	exec(t, db,
		`INSERT INTO openmos.running_orders (ro_id) VALUES ('`+strings.Repeat("x", 128)+`')`,
		`UPDATE openmos.items SET ed_dur = 4294967295, time_base = 59.94 WHERE item_id = 'i1'`,
	)
	var tb string
	if err := db.QueryRowContext(t.Context(),
		`SELECT time_base::text FROM openmos.items WHERE item_id = 'i1'`).Scan(&tb); err != nil {
		t.Fatal(err)
	}
	if tb != "59.94" {
		t.Errorf("time_base = %s, want 59.94", tb)
	}

	// A float column would print 59.94 too, so check the type is exact
	if n := count(t, db, `SELECT count(*) FROM information_schema.columns
		WHERE table_schema = 'openmos' AND column_name = 'time_base' AND data_type = 'numeric'`); n != 2 {
		t.Errorf("%d numeric time_base columns, want 2 (items, mos_objects)", n)
	}
}
