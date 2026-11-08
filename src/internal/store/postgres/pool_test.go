package postgres

import (
	"crypto/tls"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// clearPGEnv unsets the PG* variables pgx reads, so the developer's own
// environment cannot change the result.
func clearPGEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"PGHOST", "PGPORT", "PGDATABASE", "PGUSER", "PGPASSWORD",
		"PGSERVICE", "PGSSLMODE", "PGSSLROOTCERT", "PGCONNECT_TIMEOUT", "PGSSLSNI"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
}

func TestParsePoolConfigDefaults(t *testing.T) {
	clearPGEnv(t)
	cfg, err := ParsePoolConfig(PoolOptions{URL: "postgres://u@db.example/openmos?sslmode=verify-full"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxConns != DefaultMaxConns || cfg.MinConns != DefaultMinConns ||
		cfg.HealthCheckPeriod != DefaultHealthCheckPeriod || cfg.ConnConfig.ConnectTimeout != DefaultConnectTimeout {
		t.Errorf("defaults not applied: max %d, min %d, health %v, timeout %v",
			cfg.MaxConns, cfg.MinConns, cfg.HealthCheckPeriod, cfg.ConnConfig.ConnectTimeout)
	}

	// Options override the defaults, and connect_timeout in the URL wins
	cfg, err = ParsePoolConfig(PoolOptions{
		URL:      "postgres://u@db.example/openmos?sslmode=verify-full&connect_timeout=2",
		MaxConns: 4, MinConns: 2, ConnectTimeout: 9 * time.Second, HealthCheckPeriod: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxConns != 4 || cfg.MinConns != 2 || cfg.HealthCheckPeriod != time.Minute ||
		cfg.ConnConfig.ConnectTimeout != 2*time.Second {
		t.Errorf("options not applied: max %d, min %d, health %v, timeout %v",
			cfg.MaxConns, cfg.MinConns, cfg.HealthCheckPeriod, cfg.ConnConfig.ConnectTimeout)
	}

	if _, err := ParsePoolConfig(PoolOptions{URL: "postgres://u@db.example/openmos?sslmode=verify-full",
		MaxConns: 2, MinConns: 3}); err == nil {
		t.Error("min conns above max conns accepted")
	}
	for _, o := range []PoolOptions{{MaxConns: -1}, {MinConns: -1}} {
		o.URL = "postgres://u@db.example/openmos?sslmode=verify-full"
		if _, err := ParsePoolConfig(o); err == nil {
			t.Errorf("negative conns accepted: max %d, min %d", o.MaxConns, o.MinConns)
		}
	}
}

func TestParsePoolConfigRequiresVerifyFull(t *testing.T) {
	for _, tc := range []struct {
		name, url, pgsslmode string
		ok                   bool
	}{
		{"verify-full", "postgres://u@db.example/openmos?sslmode=verify-full", "", true},
		{"verify-full from PGSSLMODE", "postgres://u@db.example/openmos", "verify-full", true},
		{"default is prefer", "postgres://u@db.example/openmos", "", false},
		{"disable", "postgres://u@db.example/openmos?sslmode=disable", "", false},
		{"allow", "postgres://u@db.example/openmos?sslmode=allow", "", false},
		{"prefer", "postgres://u@db.example/openmos?sslmode=prefer", "", false},
		{"require", "postgres://u@db.example/openmos?sslmode=require", "", false},
		{"verify-ca", "postgres://u@db.example/openmos?sslmode=verify-ca", "", false},
		{"disable from PGSSLMODE", "postgres://u@db.example/openmos", "disable", false},
		{"keyword form", "host=db.example user=u sslmode=require", "", false},
		{"multiple hosts verify-full", "postgres://u@a.example,b.example/openmos?sslmode=verify-full", "", true},
		{"multiple hosts prefer", "postgres://u@a.example,b.example/openmos?sslmode=prefer", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clearPGEnv(t)
			if tc.pgsslmode != "" {
				t.Setenv("PGSSLMODE", tc.pgsslmode)
			}
			_, err := ParsePoolConfig(PoolOptions{URL: tc.url})
			if tc.ok && err != nil {
				t.Errorf("rejected: %v", err)
			}
			if !tc.ok && err == nil {
				t.Error("accepted without verify-full")
			}
			// Development may use anything
			if _, err := ParsePoolConfig(PoolOptions{URL: tc.url, AllowInsecureTLS: true}); err != nil {
				t.Errorf("rejected with AllowInsecureTLS: %v", err)
			}
		})
	}
}

func TestParseErrorWithholdsPassword(t *testing.T) {
	clearPGEnv(t)
	const secret = "s3cret-Passw0rd"
	for _, bad := range []string{
		"postgres://u:" + secret + "@db.example:notaport/openmos",
		"host=db.example password=" + secret + " port=notaport",
		"postgres://u:" + secret + "@db.example/openmos?sslmode=bogus",
	} {
		_, err := ParsePoolConfig(PoolOptions{URL: bad, AllowInsecureTLS: true})
		if err == nil {
			t.Errorf("invalid settings accepted")
			continue
		}
		// pgx redacts only on a best-effort basis, so its error is never passed on
		if !errors.Is(err, errInvalidSettings) {
			t.Errorf("err = %v, want errInvalidSettings", err)
		}
		if strings.Contains(err.Error(), secret) {
			t.Errorf("error reveals the password: %v", err)
		}
	}
}

// A single sslmode applies to every host, so this case cannot come from
// ParseConfig; the check still guards against an insecure fallback.
func TestRequireVerifyFullChecksFallbacks(t *testing.T) {
	secure := &tls.Config{ServerName: "a.example"}
	c := &pgx.ConnConfig{}
	c.TLSConfig = secure
	c.Fallbacks = []*pgconn.FallbackConfig{{Host: "b.example", TLSConfig: secure}}
	if err := requireVerifyFull(c); err != nil {
		t.Errorf("secure fallbacks rejected: %v", err)
	}
	c.Fallbacks = append(c.Fallbacks, &pgconn.FallbackConfig{Host: "c.example"})
	if err := requireVerifyFull(c); err == nil {
		t.Error("plain-text fallback accepted")
	}
}

func TestConfigured(t *testing.T) {
	clearPGEnv(t)
	if Configured("") {
		t.Error("Configured with nothing set")
	}
	if !Configured("postgres://localhost/openmos") {
		t.Error("not Configured with a URL")
	}
	t.Setenv("PGHOST", "localhost")
	if !Configured("") {
		t.Error("not Configured with PGHOST")
	}
}

func TestNewPoolConnects(t *testing.T) {
	testURL := testDatabaseURL(t)
	pool, err := NewPool(t.Context(), PoolOptions{URL: testURL, MaxConns: 3, AllowInsecureTLS: true})
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(t.Context()); err != nil {
		t.Errorf("Ping: %v", err)
	}
	if got := pool.Stat().MaxConns(); got != 3 {
		t.Errorf("MaxConns = %d, want 3", got)
	}
}

func TestNewPoolErrorWithholdsPassword(t *testing.T) {
	u, err := url.Parse(testDatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	const wrong = "wr0ng-Passw0rd"
	u.User = url.UserPassword(u.User.Username(), wrong)

	_, err = NewPool(t.Context(), PoolOptions{URL: u.String(), ConnectTimeout: 3 * time.Second, AllowInsecureTLS: true})
	if err == nil {
		t.Fatal("NewPool with a wrong password succeeded")
	}
	// Make sure the server checked the password, not just refused a dial
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != invalidPassword {
		t.Fatalf("err = %v, want SQLSTATE %s", err, invalidPassword)
	}
	if strings.Contains(err.Error(), wrong) {
		t.Errorf("error reveals the password: %v", err)
	}
}

// invalidPassword is the SQLSTATE for a failed password authentication.
const invalidPassword = "28P01"
