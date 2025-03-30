package logger

import (
	"context"
	"fmt"
	"time"

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
		return fmt.Errorf("Sentry DSN is required")
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

// SentryLogger enhances Logger with Sentry integration
type SentryLogger struct {
	*Logger
	environment string
	release     string
}

// NewSentryLogger creates a new logger with Sentry integration
func NewSentryLogger(logger *Logger, environment, release string) *SentryLogger {
	return &SentryLogger{
		Logger:      logger,
		environment: environment,
		release:     release,
	}
}

// CaptureException sends an error to Sentry
func (l *SentryLogger) CaptureException(err error, tags map[string]string, extras map[string]interface{}) {
	if err == nil {
		return
	}

	// Log the error locally first
	l.Error(err)

	// Add tags
	if tags != nil && len(tags) > 0 {
		sentry.WithScope(func(scope *sentry.Scope) {
			// Add tags
			for k, v := range tags {
				scope.SetTag(k, v)
			}

			// Add extra context
			if extras != nil && len(extras) > 0 {
				for k, v := range extras {
					scope.SetExtra(k, v)
				}
			}

			// Set environment and release if available
			if l.environment != "" {
				scope.SetTag("environment", l.environment)
			}

			if l.release != "" {
				scope.SetTag("release", l.release)
			}

			// Capture the exception
			sentry.CaptureException(err)
		})
	} else {
		// No tags, capture directly
		sentry.CaptureException(err)
	}
}

// CaptureMessage sends a message to Sentry
func (l *SentryLogger) CaptureMessage(message string, level sentry.Level, tags map[string]string, extras map[string]interface{}) {
	// Log the message locally first
	switch level {
	case sentry.LevelDebug:
		l.Debug(message)
	case sentry.LevelInfo:
		l.Info(message)
	case sentry.LevelWarning:
		l.Warning(message)
	case sentry.LevelError:
		l.Error(message)
	case sentry.LevelFatal:
		l.Error(message) // Don't call Fatal to avoid program termination
	}

	// Capture with Sentry
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)

		// Add tags
		if tags != nil && len(tags) > 0 {
			for k, v := range tags {
				scope.SetTag(k, v)
			}
		}

		// Add extra context
		if extras != nil && len(extras) > 0 {
			for k, v := range extras {
				scope.SetExtra(k, v)
			}
		}

		// Set environment and release if available
		if l.environment != "" {
			scope.SetTag("environment", l.environment)
		}

		if l.release != "" {
			scope.SetTag("release", l.release)
		}

		// Capture the message
		sentry.CaptureMessage(message)
	})
}

// StartTransaction starts a new Sentry transaction
func (l *SentryLogger) StartTransaction(name string, operation string) *sentry.Span {
	// Create the context
	ctx := context.Background()

	// Create a span (transaction)
	span := sentry.StartSpan(ctx, operation)
	span.SetTag("transaction", name)

	// Set additional info
	if l.environment != "" {
		span.SetTag("environment", l.environment)
	}

	if l.release != "" {
		span.SetTag("release", l.release)
	}

	return span
}

// Flush waits until the Sentry client has sent all events
func (l *SentryLogger) Flush(timeout time.Duration) {
	sentry.Flush(timeout)
}

// WithScope executes a function with a new scope
func (l *SentryLogger) WithScope(f func(scope *sentry.Scope)) {
	sentry.WithScope(f)
}

// ConfigureSentry configures Sentry and returns a SentryLogger
func ConfigureSentry(logger *Logger, config SentryConfig) (*SentryLogger, error) {
	err := InitializeSentry(config)
	if err != nil {
		return nil, err
	}

	return NewSentryLogger(logger, config.Environment, config.Release), nil
}
