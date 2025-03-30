package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
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
	yamlLoaded := false
	configPaths := []string{
		"config.yaml",                                // Current directory
		"config.yml",                                 // Current directory, alternative extension
		filepath.Join("config", "config.yaml"),       // Config subdirectory
		filepath.Join("config", "config.yml"),        // Config subdirectory, alternative extension
		filepath.Join("..", "config.yaml"),           // Parent directory
		filepath.Join("..", "config.yml"),            // Parent directory, alternative extension
		filepath.Join("..", "config", "config.yaml"), // Parent's config subdirectory
		filepath.Join("..", "config", "config.yml"),  // Parent's config subdirectory, alternative extension
	}

	// Also check if CONFIG_FILE env var is set
	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		configPaths = append([]string{configFile}, configPaths...)
	}

	// Try each path
	for _, path := range configPaths {
		err := loadYAMLConfig(config, path)
		if err == nil {
			// Successfully loaded
			yamlLoaded = true
			break
		} else if !os.IsNotExist(err) {
			// If error is not just "file doesn't exist", return it
			return nil, fmt.Errorf("error loading YAML config from %s: %w", path, err)
		}
	}

	// Then load from environment variables, overriding YAML values if present
	// App config
	if envVal := getEnv("APP_NAME", ""); envVal != "" || !yamlLoaded {
		config.App.Name = getEnv("APP_NAME", getDefaultString(config.App.Name, "OpenMOS"))
	}
	if envVal := getEnv("APP_VERSION", ""); envVal != "" || !yamlLoaded {
		config.App.Version = getEnv("APP_VERSION", getDefaultString(config.App.Version, "1.0.0"))
	}
	if envVal := getEnv("APP_ENV", ""); envVal != "" || !yamlLoaded {
		config.App.Environment = getEnv("APP_ENV", getDefaultString(config.App.Environment, "development"))
	}

	// Server config
	if envVal := getEnv("SERVER_HOST", ""); envVal != "" || !yamlLoaded {
		config.Server.Host = getEnv("SERVER_HOST", getDefaultString(config.Server.Host, "0.0.0.0"))
	}
	if envVal := getEnv("SERVER_PORT", ""); envVal != "" || !yamlLoaded {
		config.Server.Port = getEnvAsInt("SERVER_PORT", getDefaultInt(config.Server.Port, 10540)) // Default MOS port
	}
	if envVal := getEnv("SERVER_READ_TIMEOUT", ""); envVal != "" || !yamlLoaded {
		config.Server.ReadTimeout = getEnvAsDuration("SERVER_READ_TIMEOUT", getDefaultDuration(config.Server.ReadTimeout, 5*time.Second))
	}
	if envVal := getEnv("SERVER_WRITE_TIMEOUT", ""); envVal != "" || !yamlLoaded {
		config.Server.WriteTimeout = getEnvAsDuration("SERVER_WRITE_TIMEOUT", getDefaultDuration(config.Server.WriteTimeout, 5*time.Second))
	}
	if envVal := getEnv("SERVER_SHUTDOWN_TIMEOUT", ""); envVal != "" || !yamlLoaded {
		config.Server.ShutdownTimeout = getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", getDefaultDuration(config.Server.ShutdownTimeout, 30*time.Second))
	}

	// MongoDB config
	if envVal := getEnv("MONGODB_URI", ""); envVal != "" || !yamlLoaded {
		config.Mongo.URI = getEnv("MONGODB_URI", getDefaultString(config.Mongo.URI, "mongodb://localhost:27017"))
	}
	if envVal := getEnv("MONGODB_DATABASE", ""); envVal != "" || !yamlLoaded {
		config.Mongo.Database = getEnv("MONGODB_DATABASE", getDefaultString(config.Mongo.Database, "openmos"))
	}
	if envVal := getEnv("MONGODB_TIMEOUT", ""); envVal != "" || !yamlLoaded {
		config.Mongo.Timeout = getEnvAsDuration("MONGODB_TIMEOUT", getDefaultDuration(config.Mongo.Timeout, 10*time.Second))
	}

	// MOS config
	if envVal := getEnv("MOS_ID", ""); envVal != "" || !yamlLoaded {
		config.MOS.ID = getEnv("MOS_ID", getDefaultString(config.MOS.ID, "OpenMOS_Server"))
	}
	if envVal := getEnv("MOS_HEARTBEAT_INTERVAL", ""); envVal != "" || !yamlLoaded {
		config.MOS.HeartbeatInterval = getEnvAsDuration("MOS_HEARTBEAT_INTERVAL", getDefaultDuration(config.MOS.HeartbeatInterval, 30*time.Second))
	}
	if envVal := getEnv("MOS_CLIENT_TIMEOUT", ""); envVal != "" || !yamlLoaded {
		config.MOS.ClientTimeout = getEnvAsDuration("MOS_CLIENT_TIMEOUT", getDefaultDuration(config.MOS.ClientTimeout, 2*time.Minute))
	}

	// Logging config
	if envVal := getEnv("LOG_LEVEL", ""); envVal != "" || !yamlLoaded {
		config.Logging.Level = getEnv("LOG_LEVEL", getDefaultString(config.Logging.Level, "info"))
	}

	// Sentry config
	if envVal := getEnv("SENTRY_DSN", ""); envVal != "" || !yamlLoaded {
		config.Sentry.DSN = getEnv("SENTRY_DSN", getDefaultString(config.Sentry.DSN, ""))
	}
	if envVal := getEnv("SENTRY_ENV", ""); envVal != "" || !yamlLoaded {
		config.Sentry.Environment = getEnv("SENTRY_ENV", getDefaultString(config.Sentry.Environment, config.App.Environment))
	}
	if envVal := getEnv("SENTRY_DEBUG", ""); envVal != "" || !yamlLoaded {
		config.Sentry.Debug = getEnvAsBool("SENTRY_DEBUG", getDefaultBool(config.Sentry.Debug, false))
	}
	if envVal := getEnv("SENTRY_ATTACH_STACKTRACE", ""); envVal != "" || !yamlLoaded {
		config.Sentry.AttachStacktrace = getEnvAsBool("SENTRY_ATTACH_STACKTRACE", getDefaultBool(config.Sentry.AttachStacktrace, true))
	}
	if envVal := getEnv("SENTRY_SAMPLE_RATE", ""); envVal != "" || !yamlLoaded {
		config.Sentry.SampleRate = getEnvAsFloat("SENTRY_SAMPLE_RATE", getDefaultFloat(config.Sentry.SampleRate, 1.0))
	}
	if envVal := getEnv("SENTRY_TRACES_SAMPLE_RATE", ""); envVal != "" || !yamlLoaded {
		config.Sentry.TracesSampleRate = getEnvAsFloat("SENTRY_TRACES_SAMPLE_RATE", getDefaultFloat(config.Sentry.TracesSampleRate, 0.2))
	}

	return config, nil
}

// loadYAMLConfig loads configuration from a YAML file
func loadYAMLConfig(config *Config, filePath string) error {
	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Parse YAML
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}

	return nil
}

// SaveConfigToYAML saves the current configuration to a YAML file
func SaveConfigToYAML(config *Config, filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory for config file: %w", err)
		}
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write to file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config to file: %w", err)
	}

	return nil
}

// GenerateDefaultConfig generates a default configuration file at the specified path
func GenerateDefaultConfig(filePath string) error {
	// Create a default config
	config := &Config{}

	// App config
	config.App.Name = "OpenMOS"
	config.App.Version = "1.0.0"
	config.App.Environment = "development"

	// Server config
	config.Server.Host = "0.0.0.0"
	config.Server.Port = 10540 // Default MOS port
	config.Server.ReadTimeout = 5 * time.Second
	config.Server.WriteTimeout = 5 * time.Second
	config.Server.ShutdownTimeout = 30 * time.Second

	// MongoDB config
	config.Mongo.URI = "mongodb://localhost:27017"
	config.Mongo.Database = "openmos"
	config.Mongo.Timeout = 10 * time.Second

	// MOS config
	config.MOS.ID = "OpenMOS_Server"
	config.MOS.HeartbeatInterval = 30 * time.Second
	config.MOS.ClientTimeout = 2 * time.Minute

	// Logging config
	config.Logging.Level = "info"

	// Sentry config
	config.Sentry.DSN = "" // Empty by default
	config.Sentry.Environment = config.App.Environment
	config.Sentry.Debug = false
	config.Sentry.AttachStacktrace = true
	config.Sentry.SampleRate = 1.0
	config.Sentry.TracesSampleRate = 0.2

	// Save to file
	return SaveConfigToYAML(config, filePath)
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
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
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

// Default value helpers
func getDefaultString(current, defaultValue string) string {
	if current == "" {
		return defaultValue
	}
	return current
}

func getDefaultInt(current, defaultValue int) int {
	if current == 0 {
		return defaultValue
	}
	return current
}

func getDefaultBool(current, defaultValue bool) bool {
	// Zero value for bool is false, but can't determine if it's default without context
	// We'll just use the provided default
	return defaultValue
}

func getDefaultFloat(current, defaultValue float64) float64 {
	if current == 0 {
		return defaultValue
	}
	return current
}

func getDefaultDuration(current, defaultValue time.Duration) time.Duration {
	if current == 0 {
		return defaultValue
	}
	return current
}

// GetServerAddress returns the full server address string
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
