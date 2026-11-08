package logger

import (
	"errors"
	"fmt"

	"github.com/getsentry/sentry-go"
)

// SentryConfig holds Sentry configuration
type SentryConfig struct {
	DSN              string
	Environment      string
	Release          string
	Debug            bool
	AttachStacktrace bool
	SampleRate       float64
	TracesSampleRate float64
	ServerName       string
}

// DefaultSentryConfig returns a default Sentry configuration
func DefaultSentryConfig() SentryConfig {
	return SentryConfig{
		Environment:      "development",
		Debug:            false,
		AttachStacktrace: true,
		SampleRate:       1.0,
		TracesSampleRate: 0.2,
	}
}

// InitializeSentry initializes the Sentry client
func InitializeSentry(config SentryConfig) error {
	if config.DSN == "" {
		return errors.New("sentry DSN is required")
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.DSN,
		Environment:      config.Environment,
		Release:          config.Release,
		Debug:            config.Debug,
		AttachStacktrace: config.AttachStacktrace,
		SampleRate:       config.SampleRate,
		TracesSampleRate: config.TracesSampleRate,
		ServerName:       config.ServerName,
	})

	if err != nil {
		return fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	return nil
}

// ConfigureSentry configures Sentry and returns a SentryLogger
func ConfigureSentry(logger *Logger, config SentryConfig) (*SentryLogger, error) {
	err := InitializeSentry(config)
	if err != nil {
		return nil, err
	}

	return NewSentryLogger(logger, config.Environment, config.Release), nil
}
