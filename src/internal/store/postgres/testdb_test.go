package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
)

// testDBEnv names a throwaway database for tests. The tests drop and
// recreate the openmos schema in it, so never point it at real data.
const testDBEnv = "OPENMOS_TEST_DATABASE_URL"

// testDB returns a connection to an empty test database, or skips the test
// if testDBEnv is not set. The openmos schema is dropped before and after.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv(testDBEnv)
	if url == "" {
		// CI provides a database, so a missing URL there is a broken setup,
		// not a reason to skip.
		if os.Getenv("CI") != "" {
			t.Fatalf("%s must be set in CI", testDBEnv)
		}
		t.Skipf("%s not set; skipping PostgreSQL test", testDBEnv)
	}
	db, err := Open(url)
	if err != nil {
		t.Fatal(err)
	}
	dropSchema := func() {
		if _, err := db.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+Schema+" CASCADE"); err != nil {
			t.Fatalf("drop schema: %v", err)
		}
	}
	dropSchema()
	t.Cleanup(func() {
		dropSchema()
		db.Close()
	})
	return db
}

// migrated returns a test database with all migrations applied.
func migrated(t *testing.T) *sql.DB {
	t.Helper()
	db := testDB(t)
	if _, err := Migrate(t.Context(), db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}

// exec runs statements that must succeed.
func exec(t *testing.T, db *sql.DB, stmts ...string) {
	t.Helper()
	for _, s := range stmts {
		if _, err := db.ExecContext(t.Context(), s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
}

// count returns the result of a COUNT(*) query.
func count(t *testing.T, db *sql.DB, query string) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(t.Context(), query).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}
