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

	// MongoDB configuration
	Mongo struct {
		URI      string
		Database string
		Timeout  time.Duration
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
