package config

import "time"

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
