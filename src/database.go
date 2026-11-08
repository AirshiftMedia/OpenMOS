package main

import (
	"context"

	"airshift/openmos/internal/config"
	"airshift/openmos/internal/store/postgres"
	"airshift/openmos/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// connectDatabase opens the PostgreSQL pool if a database is configured
// (DATABASE_URL, postgres.url or the PG* variables) and exits the process if
// it cannot connect. It returns nil when no database is configured.
//
// The repositories still use the in-memory store until MOS-16. The pool is
// opened now so that bad settings or credentials fail at startup.
func connectDatabase(ctx context.Context, cfg *config.Config, log *logger.SentryLogger) *pgxpool.Pool {
	if !postgres.Configured(cfg.Postgres.URL) {
		log.Info("No PostgreSQL configured (DATABASE_URL or PG* variables); skipping database connection")
		return nil
	}

	pool, err := postgres.NewPool(ctx, postgres.PoolOptions{
		URL:               cfg.Postgres.URL,
		MaxConns:          cfg.Postgres.MaxConns,
		MinConns:          cfg.Postgres.MinConns,
		ConnectTimeout:    cfg.Postgres.ConnectTimeout,
		HealthCheckPeriod: cfg.Postgres.HealthCheckPeriod,
		// Only local development may connect without verify-full
		AllowInsecureTLS: cfg.App.Environment == "development",
	})
	if err != nil {
		// The error names host, database and user, never the password
		log.CaptureException(err, map[string]string{
			"component": "database",
			"action":    "connect",
		}, nil)
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Infof("Connected to PostgreSQL %s", postgres.Describe(pool.Config()))
	return pool
}
