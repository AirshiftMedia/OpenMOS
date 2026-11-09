package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// DATABASE_URL (the backend01 .env name) overrides postgres.url in YAML,
// and the other postgres fields come from YAML.
func TestDatabaseURLOverridesYAML(t *testing.T) {
	dir := isolate(t)
	path := filepath.Join(dir, "config.yaml")
	yamlDoc := "postgres:\n    url: postgres://from-yaml/openmos\n    migrateurl: postgres://migrate-yaml/openmos\n    maxconns: 4\n    connecttimeout: 3s\n"
	if err := os.WriteFile(path, []byte(yamlDoc), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Postgres.URL != "postgres://from-yaml/openmos" || cfg.Postgres.MigrateURL != "postgres://migrate-yaml/openmos" ||
		cfg.Postgres.MaxConns != 4 ||
		cfg.Postgres.ConnectTimeout != 3*time.Second {
		t.Errorf("YAML postgres block not loaded: %+v", cfg.Postgres)
	}

	t.Setenv("DATABASE_URL", "postgres://from-env/openmos")
	t.Setenv("MIGRATE_DATABASE_URL", "postgres://migrate-env/openmos")
	t.Setenv("POSTGRES_MAX_CONNS", "7")
	cfg, err = LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Postgres.URL != "postgres://from-env/openmos" || cfg.Postgres.MigrateURL != "postgres://migrate-env/openmos" ||
		cfg.Postgres.MaxConns != 7 {
		t.Errorf("env did not override: URL %q, MigrateURL %q, MaxConns %d",
			cfg.Postgres.URL, cfg.Postgres.MigrateURL, cfg.Postgres.MaxConns)
	}
}

// The generated default config never contains a database URL.
func TestGeneratedConfigHasNoDatabaseURL(t *testing.T) {
	dir := isolate(t)
	path := filepath.Join(dir, "generated.yaml")
	if err := GenerateDefaultConfig(path); err != nil {
		t.Fatalf("GenerateDefaultConfig: %v", err)
	}
	t.Setenv("CONFIG_FILE", path)
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Postgres.URL != "" || cfg.Postgres.MigrateURL != "" || cfg.Postgres.MaxConns != 10 {
		t.Errorf("generated postgres block = %+v, want empty URL and defaults", cfg.Postgres)
	}
}

// Pool sizes that do not fit in an int32 fall back to the default instead
// of wrapping around.
func TestPostgresConnsOutOfRangeFallBack(t *testing.T) {
	for _, v := range []string{"2147483648", "99999999999", "-2147483649", "ten"} {
		isolate(t)
		t.Setenv("POSTGRES_MAX_CONNS", v)
		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if cfg.Postgres.MaxConns != 10 {
			t.Errorf("POSTGRES_MAX_CONNS=%s: MaxConns = %d, want default 10", v, cfg.Postgres.MaxConns)
		}
	}
	isolate(t)
	t.Setenv("POSTGRES_MAX_CONNS", "2147483647")
	if cfg, err := LoadConfig(); err != nil || cfg.Postgres.MaxConns != 2147483647 {
		t.Errorf("largest int32 not accepted: %v, %v", cfg.Postgres.MaxConns, err)
	}
}
