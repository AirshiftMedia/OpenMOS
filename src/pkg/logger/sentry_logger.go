package logger

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
)

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
