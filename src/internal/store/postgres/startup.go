package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// ErrSchemaOutdated means the database schema is older than this build
// needs: pending migrations must be applied first.
var ErrSchemaOutdated = errors.New("database schema is outdated")

// MigrateWith connects with opts, applies all pending migrations and
// returns the versions applied. The connection gets the same checks as
// NewPool (verify-full outside development) and is closed afterwards. Use
// the schema owner over a direct connection (openmos_migrate on Aiven).
func MigrateWith(ctx context.Context, opts PoolOptions) ([]int64, error) {
	cfg, err := ParsePoolConfig(opts)
	if err != nil {
		return nil, err
	}
	db := stdlib.OpenDB(*cfg.ConnConfig)
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnConfig.ConnectTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("connect to PostgreSQL %s for migrations: %w", Describe(cfg), err)
	}
	applied, err := Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate %s: %w", Describe(cfg), err)
	}
	return applied, nil
}

// CheckSchema returns the schema version of the database behind pool and
// the latest version this build has migrations for. It returns
// ErrSchemaOutdated if migrations are pending, including on a database
// that was never migrated. A newer schema than the build knows, as after
// rolling back a deploy, is not an error.
func CheckSchema(ctx context.Context, pool *pgxpool.Pool) (current, latest int64, err error) {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	var tableExists bool
	err = db.QueryRowContext(ctx, "SELECT to_regclass($1) IS NOT NULL", Schema+".goose_db_version").Scan(&tableExists)
	if err != nil {
		return 0, 0, fmt.Errorf("check schema version: %w", err)
	}
	p, err := newProvider(db)
	if err != nil {
		return 0, 0, err
	}
	sources := p.ListSources()
	if len(sources) > 0 {
		latest = sources[len(sources)-1].Version
	}
	if tableExists {
		if current, err = p.GetDBVersion(ctx); err != nil {
			return 0, 0, fmt.Errorf("check schema version: %w", err)
		}
	}
	if current < latest {
		return current, latest, fmt.Errorf("schema at version %d, this build needs %d: %w",
			current, latest, ErrSchemaOutdated)
	}
	return current, latest, nil
}
