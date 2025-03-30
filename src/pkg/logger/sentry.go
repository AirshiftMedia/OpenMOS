package logger

import (
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

	// Create a Sentry event
	event := sentry.NewEvent()
	event.Level = sentry.LevelError
	event.Message = err.Error()

	// Add environment and release if available
	if l.environment != "" {
		event.Environment = l.environment
	}

	if l.release != "" {
		event.Release = l.release
	}

	// Add tags
	if tags != nil && len(tags) > 0 {
		for k, v := range tags {
			event.Tags[k] = v
		}
	}

	// Add extra context
	if extras != nil && len(extras) > 0 {
		for k, v := range extras {
			event.Extra[k] = v
		}
	}

	// Capture the exception
	sentry.CaptureEvent(event)
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

	// Create a Sentry event
	event := sentry.NewEvent()
	event.Level = level
	event.Message = message

	// Add environment and release if available
	if l.environment != "" {
		event.Environment = l.environment
	}

	if l.release != "" {
		event.Release = l.release
	}

	// Add tags
	if tags != nil && len(tags) > 0 {
		for k, v := range tags {
			event.Tags[k] = v
		}
	}

	// Add extra context
	if extras != nil && len(extras) > 0 {
		for k, v := range extras {
			event.Extra[k] = v
		}
	}

	// Capture the message
	sentry.CaptureEvent(event)
}

// StartTransaction starts a new Sentry transaction
func (l *SentryLogger) StartTransaction(name, operation string) *sentry.Transaction {
	return sentry.StartTransaction(
		sentry.Context{
			Name: name,
			Op:   operation,
		},
	)
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
