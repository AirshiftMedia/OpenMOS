package postgres

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testPool returns a pool on the test database, closed at the end of the
// test.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(t.Context(), testDatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func devOptions(t *testing.T) PoolOptions {
	return PoolOptions{URL: testDatabaseURL(t), AllowInsecureTLS: true}
}

func TestMigrateWith(t *testing.T) {
	db := testDB(t) // empty database, no schema
	ctx := t.Context()

	applied, err := MigrateWith(ctx, devOptions(t))
	if err != nil {
		t.Fatalf("MigrateWith: %v", err)
	}
	if !slices.Contains(applied, 1) {
		t.Errorf("first MigrateWith applied %v, want version 1 among them", applied)
	}
	if v, err := Version(ctx, db); err != nil || v != applied[len(applied)-1] {
		t.Errorf("Version after MigrateWith = %d, %v; want %d", v, err, applied[len(applied)-1])
	}

	applied, err = MigrateWith(ctx, devOptions(t))
	if err != nil || len(applied) != 0 {
		t.Errorf("second MigrateWith applied %v, %v; want nothing", applied, err)
	}
}

// The migration connection has the same TLS rule as the pool, and fails
// before connecting
func TestMigrateWithRequiresVerifyFull(t *testing.T) {
	url := testDatabaseURL(t)
	if !strings.Contains(url, "sslmode=disable") {
		t.Skip("test database URL does not use sslmode=disable")
	}
	_, err := MigrateWith(t.Context(), PoolOptions{URL: url})
	if err == nil || !strings.Contains(err.Error(), "verify-full") {
		t.Errorf("MigrateWith without verify-full: err = %v, want the verify-full rule", err)
	}
}

func TestMigrateWithBadPassword(t *testing.T) {
	url := testDatabaseURL(t)
	bad := strings.Replace(url, ":openmos@", ":wrong-secret@", 1)
	if bad == url {
		t.Skip("test database URL does not have the expected openmos password")
	}
	_, err := MigrateWith(t.Context(), PoolOptions{URL: bad, AllowInsecureTLS: true})
	var pe *pgconn.PgError
	if !errors.As(err, &pe) || pe.Code != "28P01" {
		t.Fatalf("MigrateWith with a wrong password: err = %v, want SQLSTATE 28P01", err)
	}
	if strings.Contains(err.Error(), "wrong-secret") {
		t.Errorf("error reveals the password: %v", err)
	}
}

func TestCheckSchema(t *testing.T) {
	db := testDB(t)
	pool := testPool(t)
	ctx := t.Context()

	// Never migrated: no schema at all
	current, latest, err := CheckSchema(ctx, pool)
	if !errors.Is(err, ErrSchemaOutdated) || current != 0 || latest < 1 {
		t.Fatalf("CheckSchema on empty database = %d, %d, %v; want 0, >=1, %v", current, latest, err, ErrSchemaOutdated)
	}

	if _, err := Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	current, latest, err = CheckSchema(ctx, pool)
	if err != nil || current != latest {
		t.Errorf("CheckSchema after Migrate = %d, %d, %v; want equal versions, nil", current, latest, err)
	}

	// Rolled back: the version table exists but is behind
	if err := MigrateDownTo(ctx, db, 0); err != nil {
		t.Fatal(err)
	}
	if _, _, err := CheckSchema(ctx, pool); !errors.Is(err, ErrSchemaOutdated) {
		t.Errorf("CheckSchema after rollback: err = %v, want %v", err, ErrSchemaOutdated)
	}

	// A schema newer than the build is accepted
	if _, err := Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	exec(t, db, "INSERT INTO openmos.goose_db_version (version_id, is_applied) VALUES (99999, true)")
	current, latest, err = CheckSchema(ctx, pool)
	if err != nil || current != 99999 || latest >= current {
		t.Errorf("CheckSchema with newer schema = %d, %d, %v; want 99999, older, nil", current, latest, err)
	}
}
