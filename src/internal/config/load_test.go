package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var configEnvKeys = []string{
	"CONFIG_FILE",
	"APP_NAME", "APP_VERSION", "APP_ENV",
	"SERVER_HOST", "SERVER_PORT", "SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT", "SERVER_SHUTDOWN_TIMEOUT",
	"MONGODB_URI", "MONGODB_DATABASE", "MONGODB_TIMEOUT",
	"MOS_ID", "MOS_HEARTBEAT_INTERVAL", "MOS_CLIENT_TIMEOUT",
	"LOG_LEVEL",
	"SENTRY_DSN", "SENTRY_ENV", "SENTRY_DEBUG", "SENTRY_ATTACH_STACKTRACE",
	"SENTRY_SAMPLE_RATE", "SENTRY_TRACES_SAMPLE_RATE",
}

// isolate unsets every config env var and runs the test from a nested temp
// directory, so neither the environment nor a stray config.yaml in the
// searched paths (., config/, .., ../config/) affects LoadConfig.
func isolate(t *testing.T) string {
	t.Helper()
	for _, k := range configEnvKeys {
		t.Setenv(k, "") // registers restore of the original value
		os.Unsetenv(k)
	}
	dir := filepath.Join(t.TempDir(), "a", "b")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	return dir
}

func TestLoadConfigDefaultsWithoutYAML(t *testing.T) {
	isolate(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	want := Config{}
	want.App.Name, want.App.Version, want.App.Environment = "OpenMOS", "1.0.0", "development"
	want.Server.Host, want.Server.Port = "0.0.0.0", 10540
	want.Server.ReadTimeout, want.Server.WriteTimeout = 5*time.Second, 5*time.Second
	want.Server.ShutdownTimeout = 30 * time.Second
	want.Mongo.URI, want.Mongo.Database = "mongodb://localhost:27017", "openmos"
	want.Mongo.Timeout = 10 * time.Second
	want.MOS.ID = "OpenMOS_Server"
	want.MOS.HeartbeatInterval, want.MOS.ClientTimeout = 30*time.Second, 2*time.Minute
	want.Logging.Level = "info"
	want.Sentry.Environment = "development"
	want.Sentry.AttachStacktrace = true
	want.Sentry.SampleRate, want.Sentry.TracesSampleRate = 1.0, 0.2

	if *cfg != want {
		t.Errorf("defaults mismatch:\n got  %+v\n want %+v", *cfg, want)
	}
}

func TestLoadConfigEnvOverridesWithoutYAML(t *testing.T) {
	isolate(t)
	t.Setenv("SERVER_PORT", "10541")
	t.Setenv("MOS_CLIENT_TIMEOUT", "45s")
	t.Setenv("SENTRY_DEBUG", "true")
	t.Setenv("APP_ENV", "staging")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Server.Port != 10541 || cfg.MOS.ClientTimeout != 45*time.Second || !cfg.Sentry.Debug {
		t.Errorf("env overrides not applied: %+v", *cfg)
	}
	// SENTRY_ENV defaults to the app environment
	if cfg.App.Environment != "staging" || cfg.Sentry.Environment != "staging" {
		t.Errorf("App.Environment = %q, Sentry.Environment = %q, want staging",
			cfg.App.Environment, cfg.Sentry.Environment)
	}
}

// With a YAML file loaded, only explicitly set env vars override it and
// fields missing from the YAML are not filled with defaults.
func TestLoadConfigFromConfigFileWithEnvOverride(t *testing.T) {
	dir := isolate(t)
	path := filepath.Join(dir, "custom.yaml")
	yamlDoc := "server:\n    port: 11000\nmos:\n    id: from.yaml\n"
	if err := os.WriteFile(path, []byte(yamlDoc), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_FILE", path)
	t.Setenv("MOS_ID", "from.env")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Server.Port != 11000 {
		t.Errorf("Server.Port = %d, want 11000 from YAML", cfg.Server.Port)
	}
	if cfg.MOS.ID != "from.env" {
		t.Errorf("MOS.ID = %q, want env override", cfg.MOS.ID)
	}
	if cfg.Server.Host != "" || cfg.Logging.Level != "" {
		t.Errorf("fields absent from YAML got defaults: host %q, level %q",
			cfg.Server.Host, cfg.Logging.Level)
	}
}

func TestLoadConfigSearchesWorkingDirectory(t *testing.T) {
	dir := isolate(t)
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("mos:\n    id: cwd\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.MOS.ID != "cwd" {
		t.Errorf("MOS.ID = %q, want cwd", cfg.MOS.ID)
	}
}

func TestLoadConfigInvalidYAMLReturnsError(t *testing.T) {
	dir := isolate(t)
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("server: [unclosed"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_FILE", path)

	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig with invalid YAML returned nil error")
	}
}

func TestLoadConfigMissingConfigFileFallsBackToDefaults(t *testing.T) {
	dir := isolate(t)
	t.Setenv("CONFIG_FILE", filepath.Join(dir, "does-not-exist.yaml"))

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.MOS.ID != "OpenMOS_Server" {
		t.Errorf("MOS.ID = %q, want default", cfg.MOS.ID)
	}
}
