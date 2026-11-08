// Package postgres is the PostgreSQL store: schema migrations now, and the
// connection pool and repositories as they are added (MOS-13, MOS-14).
package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	// Registers the "pgx" database/sql driver that goose runs on.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Schema is the PostgreSQL schema that holds all OpenMOS tables, including
// the goose version table.
const Schema = "openmos"

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Open opens a database/sql handle on the pgx driver for running
// migrations. Use a direct connection, not a PgBouncer transaction pool:
// goose holds a session-level advisory lock while it migrates.
func Open(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	return db, nil
}

// Migrate applies all pending migrations and returns the versions applied.
// It runs as the schema owner (openmos_migrate on Aiven). If the schema does
// not exist yet, as in a fresh local database, it is created first.
func Migrate(ctx context.Context, db *sql.DB) ([]int64, error) {
	if err := ensureSchema(ctx, db); err != nil {
		return nil, err
	}
	p, err := newProvider(db)
	if err != nil {
		return nil, err
	}
	results, err := p.Up(ctx)
	if err != nil {
		return nil, fmt.Errorf("migrate up: %w", err)
	}
	applied := make([]int64, len(results))
	for i, r := range results {
		applied[i] = r.Source.Version
	}
	return applied, nil
}

// MigrateDownTo rolls back migrations until version is the latest applied;
// version 0 rolls back everything. It is for tests and manual rollback.
func MigrateDownTo(ctx context.Context, db *sql.DB, version int64) error {
	p, err := newProvider(db)
	if err != nil {
		return err
	}
	if _, err := p.DownTo(ctx, version); err != nil {
		return fmt.Errorf("migrate down to %d: %w", version, err)
	}
	return nil
}

// Version returns the latest applied migration version, 0 if none.
func Version(ctx context.Context, db *sql.DB) (int64, error) {
	p, err := newProvider(db)
	if err != nil {
		return 0, err
	}
	return p.GetDBVersion(ctx)
}

func newProvider(db *sql.DB) (*goose.Provider, error) {
	fsys, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return nil, err
	}
	// The advisory lock keeps two deploys from migrating at the same time.
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, err
	}
	p, err := goose.NewProvider(goose.DialectPostgres, db, fsys,
		goose.WithTableName(Schema+".goose_db_version"),
		goose.WithSessionLocker(locker),
		goose.WithDisableGlobalRegistry(true),
	)
	if err != nil {
		return nil, fmt.Errorf("create migration provider: %w", err)
	}
	return p, nil
}

// ensureSchema creates the schema if it is missing. On Aiven it already
// exists, owned by openmos_migrate, so no CREATE privilege on the database
// is needed there.
func ensureSchema(ctx context.Context, db *sql.DB) error {
	var exists bool
	err := db.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = $1)", Schema).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check schema %s: %w", Schema, err)
	}
	if exists {
		return nil
	}
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+Schema); err != nil {
		return fmt.Errorf("create schema %s: %w", Schema, err)
	}
	return nil
}
