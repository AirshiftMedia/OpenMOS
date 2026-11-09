package config

import (
	"fmt"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Application information
	App struct {
		Name        string
		Version     string
		Environment string
	}

	// Server configuration
	Server struct {
		Host            string
		Port            int
		ReadTimeout     time.Duration
		WriteTimeout    time.Duration
		ShutdownTimeout time.Duration
	}

	// PostgreSQL configuration. URL is a secret (it holds the password): set
	// it with DATABASE_URL or in a local, gitignored config.yaml. Without a
	// URL, the standard PG* environment variables are used. MigrateURL, also
	// a secret (MIGRATE_DATABASE_URL), is used for --migrate when set, so
	// migrations can run as the schema owner over a direct connection;
	// otherwise URL is used.
	Postgres struct {
		URL               string
		MigrateURL        string
		MaxConns          int32
		MinConns          int32
		ConnectTimeout    time.Duration
		HealthCheckPeriod time.Duration
	}

	// MOS configuration
	MOS struct {
		// MOS ID of this server, used in MOS messages
		ID string
		// Heartbeat interval
		HeartbeatInterval time.Duration
		// Timeout for client connections without heartbeats
		ClientTimeout time.Duration
	}

	// Logging configuration
	Logging struct {
		Level string
	}

	// Sentry configuration
	Sentry struct {
		DSN              string
		Environment      string
		Debug            bool
		AttachStacktrace bool
		SampleRate       float64
		TracesSampleRate float64
	}
}

// LoadConfig loads configuration from environment variables and a YAML file if available
func LoadConfig() (*Config, error) {
	config := &Config{}

	// First, try to load from YAML file
	yamlLoaded, err := loadFirstYAMLConfig(config)
	if err != nil {
		return nil, err
	}

	// Then load from environment variables, overriding YAML values if present
	applyEnvOverrides(config, yamlLoaded)

	return config, nil
}

// GetServerAddress returns the full server address string
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
