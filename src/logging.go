package main

import (
	"strings"

	"airshift/openmos/internal/config"
	"airshift/openmos/pkg/logger"
)

// generateDefaultConfig writes a default configuration file to path.
func generateDefaultConfig(path string) {
	// Initialize standard logger first
	standardLogger := logger.DefaultLogger()
	standardLogger.Info("Generating default configuration file...")

	err := config.GenerateDefaultConfig(path)
	if err != nil {
		standardLogger.Fatalf("Failed to generate configuration file: %v", err)
	}

	standardLogger.Infof("Configuration file generated at: %s", path)
}

// setupLogging applies the configured log level, sets up Sentry when a DSN
// is configured, and installs the standard logger as the global logger.
func setupLogging(cfg *config.Config, standardLogger *logger.Logger) *logger.SentryLogger {
	// Configure log level
	logLevel, exists := logger.LevelValues[strings.ToLower(cfg.Logging.Level)]
	if !exists {
		standardLogger.Warningf("Unknown log level: %s. Using 'info' level.", cfg.Logging.Level)
		logLevel = logger.LevelInfo
	}
	standardLogger.SetLevel(logLevel)

	// Configure Sentry if DSN is provided
	var log *logger.SentryLogger
	if cfg.Sentry.DSN != "" {
		sentryConfig := logger.SentryConfig{
			DSN:              cfg.Sentry.DSN,
			Environment:      cfg.Sentry.Environment,
			Release:          cfg.App.Version,
			Debug:            cfg.Sentry.Debug,
			AttachStacktrace: cfg.Sentry.AttachStacktrace,
			SampleRate:       cfg.Sentry.SampleRate,
			TracesSampleRate: cfg.Sentry.TracesSampleRate,
			ServerName:       cfg.App.Name,
		}

		sentryLogger, err := logger.ConfigureSentry(standardLogger, sentryConfig)
		if err != nil {
			standardLogger.Errorf("Failed to configure Sentry: %v, continuing without Sentry integration", err)
			log = logger.NewSentryLogger(standardLogger, cfg.App.Environment, cfg.App.Version)
		} else {
			log = sentryLogger
			log.Info("Sentry integration configured successfully")
		}
	} else {
		log = logger.NewSentryLogger(standardLogger, cfg.App.Environment, cfg.App.Version)
		log.Info("Sentry DSN not provided, continuing without Sentry integration")
	}

	// Set as global logger
	logger.SetGlobalLogger(standardLogger)

	return log
}
