package config

import (
	"fmt"
	"os"
	"strconv"
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

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{}

	// App config
	config.App.Name = getEnv("APP_NAME", "OpenMOS")
	config.App.Version = getEnv("APP_VERSION", "1.0.0")
	config.App.Environment = getEnv("APP_ENV", "development")

	// Server config
	config.Server.Host = getEnv("SERVER_HOST", "0.0.0.0")
	config.Server.Port = getEnvAsInt("SERVER_PORT", 10540) // Default MOS port
	config.Server.ReadTimeout = getEnvAsDuration("SERVER_READ_TIMEOUT", 5*time.Second)
	config.Server.WriteTimeout = getEnvAsDuration("SERVER_WRITE_TIMEOUT", 5*time.Second)
	config.Server.ShutdownTimeout = getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second)

	// MongoDB config
	config.Mongo.URI = getEnv("MONGODB_URI", "mongodb://localhost:27017")
	config.Mongo.Database = getEnv("MONGODB_DATABASE", "openmos")
	config.Mongo.Timeout = getEnvAsDuration("MONGODB_TIMEOUT", 10*time.Second)

	// MOS config
	config.MOS.ID = getEnv("MOS_ID", "OpenMOS_Server")
	config.MOS.HeartbeatInterval = getEnvAsDuration("MOS_HEARTBEAT_INTERVAL", 30*time.Second)
	config.MOS.ClientTimeout = getEnvAsDuration("MOS_CLIENT_TIMEOUT", 2*time.Minute)

	// Logging config
	config.Logging.Level = getEnv("LOG_LEVEL", "info")

	// Sentry config
	config.Sentry.DSN = getEnv("SENTRY_DSN", "")
	config.Sentry.Environment = getEnv("SENTRY_ENV", config.App.Environment)
	config.Sentry.Debug = getEnvAsBool("SENTRY_DEBUG", false)
	config.Sentry.AttachStacktrace = getEnvAsBool("SENTRY_ATTACH_STACKTRACE", true)
	config.Sentry.SampleRate = getEnvAsFloat("SENTRY_SAMPLE_RATE", 1.0)
	config.Sentry.TracesSampleRate = getEnvAsFloat("SENTRY_TRACES_SAMPLE_RATE", 0.2)

	return config, nil
}

// Helper functions to get environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetServerAddress returns the full server address string
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
