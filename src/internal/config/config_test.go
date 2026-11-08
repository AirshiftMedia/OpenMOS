package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Covers the YAML library switch (MOS-8): durations and nested sections
// must decode the same way as with gopkg.in/yaml.v3.
func TestLoadYAMLConfigParsesDurationsAndSections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	yamlDoc := `server:
    host: 0.0.0.0
    port: 10540
    readtimeout: 5s
    shutdowntimeout: 30s
mos:
    id: mos01.station.com
    heartbeatinterval: 30s
    clienttimeout: 2m0s
sentry:
    attachstacktrace: true
    tracessamplerate: 0.2
`
	if err := os.WriteFile(path, []byte(yamlDoc), 0o600); err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := loadYAMLConfig(&cfg, path); err != nil {
		t.Fatalf("loadYAMLConfig: %v", err)
	}

	if cfg.Server.Port != 10540 || cfg.Server.ReadTimeout != 5*time.Second ||
		cfg.Server.ShutdownTimeout != 30*time.Second {
		t.Errorf("Server = %+v", cfg.Server)
	}
	if cfg.MOS.ID != "mos01.station.com" || cfg.MOS.HeartbeatInterval != 30*time.Second ||
		cfg.MOS.ClientTimeout != 2*time.Minute {
		t.Errorf("MOS = %+v", cfg.MOS)
	}
	if !cfg.Sentry.AttachStacktrace || cfg.Sentry.TracesSampleRate != 0.2 {
		t.Errorf("Sentry = %+v", cfg.Sentry)
	}
}

func TestSaveAndLoadConfigRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "config.yaml")

	var want Config
	want.App.Name = "OpenMOS"
	want.Server.Port = 10541
	want.Server.WriteTimeout = 7 * time.Second
	want.MOS.ID = "mos.test"
	want.MOS.ClientTimeout = 90 * time.Second
	want.Sentry.SampleRate = 0.5

	if err := SaveConfigToYAML(&want, path); err != nil {
		t.Fatalf("SaveConfigToYAML: %v", err)
	}

	var got Config
	if err := loadYAMLConfig(&got, path); err != nil {
		t.Fatalf("loadYAMLConfig: %v", err)
	}
	if got != want {
		t.Errorf("round trip mismatch:\n got  %+v\n want %+v", got, want)
	}
}
