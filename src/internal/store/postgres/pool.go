package postgres

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool defaults, used for options left at zero.
const (
	DefaultMaxConns          = 10
	DefaultMinConns          = 1
	DefaultConnectTimeout    = 5 * time.Second
	DefaultHealthCheckPeriod = 30 * time.Second
)

// PoolOptions configures the connection pool. Zero values get the defaults.
type PoolOptions struct {
	// URL is a libpq connection string or postgres:// URL. It holds the
	// password, so it is never logged. Empty means the PG* environment
	// variables alone (PGHOST, PGUSER, PGSSLMODE, ...).
	URL string

	MaxConns          int32
	MinConns          int32
	ConnectTimeout    time.Duration
	HealthCheckPeriod time.Duration

	// AllowInsecureTLS permits any sslmode. Without it, every connection
	// must use sslmode=verify-full. Only local development should set it.
	AllowInsecureTLS bool
}

// Configured reports whether a database is configured, through a URL or
// the PG* environment variables.
func Configured(url string) bool {
	if url != "" {
		return true
	}
	for _, k := range []string{"PGHOST", "PGDATABASE", "PGSERVICE"} {
		if os.Getenv(k) != "" {
			return true
		}
	}
	return false
}

// errInvalidSettings replaces pgx parse errors, which redact passwords only
// on a best-effort basis.
var errInvalidSettings = errors.New("invalid PostgreSQL connection settings (details withheld because they may contain the password)")

// ParsePoolConfig builds a pool configuration from opts and checks it. It
// does not connect.
func ParsePoolConfig(opts PoolOptions) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(opts.URL)
	if err != nil {
		return nil, errInvalidSettings
	}

	cfg.MaxConns = valueOr(opts.MaxConns, DefaultMaxConns)
	cfg.MinConns = valueOr(opts.MinConns, DefaultMinConns)
	if cfg.MinConns > cfg.MaxConns {
		return nil, fmt.Errorf("PostgreSQL min conns (%d) exceeds max conns (%d)", cfg.MinConns, cfg.MaxConns)
	}
	cfg.HealthCheckPeriod = valueOr(opts.HealthCheckPeriod, DefaultHealthCheckPeriod)
	// A connect_timeout in the URL or PGCONNECT_TIMEOUT takes precedence
	if cfg.ConnConfig.ConnectTimeout == 0 {
		cfg.ConnConfig.ConnectTimeout = valueOr(opts.ConnectTimeout, DefaultConnectTimeout)
	}

	if !opts.AllowInsecureTLS {
		if err := requireVerifyFull(cfg.ConnConfig); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// NewPool creates a connection pool and pings the database, so a bad
// configuration fails at startup. Errors never include the password.
func NewPool(ctx context.Context, opts PoolOptions) (*pgxpool.Pool, error) {
	cfg, err := ParsePoolConfig(opts)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool for %s: %w", Describe(cfg), err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnConfig.ConnectTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to PostgreSQL %s: %w", Describe(cfg), err)
	}
	return pool, nil
}

// Describe identifies the database for logs: host, port, database and
// user, never the password.
func Describe(cfg *pgxpool.Config) string {
	c := cfg.ConnConfig
	return fmt.Sprintf("%s:%d/%s (user %s)", c.Host, c.Port, c.Database, c.User)
}

// requireVerifyFull fails unless every host and fallback uses TLS with
// certificate and host name verification. pgx does not keep the sslmode,
// so it is read from the resulting TLS settings: disable has none, allow
// and prefer fall back to plain connections, and require and verify-ca
// skip host name verification (InsecureSkipVerify).
func requireVerifyFull(c *pgx.ConnConfig) error {
	configs := []*tls.Config{c.TLSConfig}
	for _, fb := range c.Fallbacks {
		configs = append(configs, fb.TLSConfig)
	}
	for _, t := range configs {
		if t == nil || t.InsecureSkipVerify {
			return errors.New("PostgreSQL connections must use sslmode=verify-full with sslrootcert " +
				"set to the server CA outside local development")
		}
	}
	return nil
}

func valueOr[T comparable](v, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
}
