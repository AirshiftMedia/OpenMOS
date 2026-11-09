package main

import (
	"context"
	"errors"

	"airshift/openmos/internal/config"
	"airshift/openmos/internal/store/postgres"
	"airshift/openmos/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// poolOptions returns the connection settings for url. Only local
// development may connect without sslmode=verify-full.
func poolOptions(cfg *config.Config, url string) postgres.PoolOptions {
	return postgres.PoolOptions{
		URL:               url,
		MaxConns:          cfg.Postgres.MaxConns,
		MinConns:          cfg.Postgres.MinConns,
		ConnectTimeout:    cfg.Postgres.ConnectTimeout,
		HealthCheckPeriod: cfg.Postgres.HealthCheckPeriod,
		AllowInsecureTLS:  cfg.App.Environment == "development",
	}
}

// migrationURL is the connection for --migrate: MIGRATE_DATABASE_URL if
// set (the schema owner over a direct connection, openmos_migrate on
// Aiven), otherwise the application's own.
func migrationURL(cfg *config.Config) string {
	if cfg.Postgres.MigrateURL != "" {
		return cfg.Postgres.MigrateURL
	}
	return cfg.Postgres.URL
}

// fatalDatabase reports a database error and exits. Errors from the
// postgres package name host, database and user, never the password.
func fatalDatabase(log *logger.SentryLogger, action string, err error, format string) {
	log.CaptureException(err, map[string]string{
		"component": "database",
		"action":    action,
	}, nil)
	log.Fatalf(format, err)
}

// migrateDatabase applies pending schema migrations, or exits the process
// if it cannot.
func migrateDatabase(ctx context.Context, cfg *config.Config, log *logger.SentryLogger) {
	url := migrationURL(cfg)
	if !postgres.Configured(url) {
		log.Fatalf("Cannot migrate: no PostgreSQL configured (set MIGRATE_DATABASE_URL, DATABASE_URL or the PG* variables)")
	}
	applied, err := postgres.MigrateWith(ctx, poolOptions(cfg, url))
	if err != nil {
		fatalDatabase(log, "migrate", err, "Failed to migrate the database schema: %v")
	}
	if len(applied) == 0 {
		log.Info("Database schema is up to date; no migrations applied")
		return
	}
	log.Infof("Applied schema migrations %v", applied)
}

// openDatabase connects to PostgreSQL and checks that the schema is
// migrated, or exits the process. The server cannot run without it.
func openDatabase(ctx context.Context, cfg *config.Config, log *logger.SentryLogger) *pgxpool.Pool {
	if !postgres.Configured(cfg.Postgres.URL) {
		log.Fatalf("No PostgreSQL configured: set DATABASE_URL (or postgres.url, or the PG* variables)")
	}
	pool, err := postgres.NewPool(ctx, poolOptions(cfg, cfg.Postgres.URL))
	if err != nil {
		fatalDatabase(log, "connect", err, "Failed to connect to PostgreSQL: %v")
	}
	log.Infof("Connected to PostgreSQL %s", postgres.Describe(pool.Config()))

	current, latest, err := postgres.CheckSchema(ctx, pool)
	switch {
	case errors.Is(err, postgres.ErrSchemaOutdated):
		pool.Close()
		log.Fatalf("%v; run with --migrate, or --migrate-only before starting", err)
	case err != nil:
		pool.Close()
		fatalDatabase(log, "check_schema", err, "Failed to check the database schema: %v")
	case current > latest:
		log.Warningf("Database schema version %d is newer than this build (%d); a newer release migrated it", current, latest)
	default:
		log.Infof("Database schema at version %d", current)
	}
	return pool
}
