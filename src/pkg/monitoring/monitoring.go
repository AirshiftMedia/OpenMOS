package monitoring

import (
	"context"
	"time"

	"airshift/openmos/pkg/logger"

	"github.com/getsentry/sentry-go"
)

// MonitoredOperation represents an operation being monitored
type MonitoredOperation struct {
	span      *sentry.Span
	name      string
	operation string
	start     time.Time
	tags      map[string]string
}

// StartOperation starts monitoring an operation
func StartOperation(ctx context.Context, name, operation string) (*MonitoredOperation, context.Context) {
	// Create a span
	span := sentry.StartSpan(ctx, operation)
	span.SetTag("transaction", name)

	// Create a new context with the span
	newCtx := sentry.SetSpanContext(ctx, span)

	return &MonitoredOperation{
		span:      span,
		name:      name,
		operation: operation,
		start:     time.Now(),
		tags:      make(map[string]string),
	}, newCtx
}

// SetTag adds a tag to the operation
func (m *MonitoredOperation) SetTag(key, value string) {
	if m == nil || m.span == nil {
		return
	}

	m.tags[key] = value
	m.span.SetTag(key, value)
}

// SetData adds data to the operation
func (m *MonitoredOperation) SetData(key string, value interface{}) {
	if m == nil || m.span == nil {
		return
	}

	m.span.SetData(key, value)
}

// Finish completes the operation
func (m *MonitoredOperation) Finish() {
	if m == nil || m.span == nil {
		return
	}

	duration := time.Since(m.start)

	// Complete the span
	m.span.Finish()

	// Log the operation for local visibility
	logger.Debugf("Operation %s (%s) completed in %v", m.name, m.operation, duration)
}

// RecordError records an error that occurred during the operation
func (m *MonitoredOperation) RecordError(err error) {
	if m == nil || m.span == nil || err == nil {
		return
	}

	// Set error status on span
	m.span.Status = sentry.SpanStatusInternalError

	// Capture exception with tags
	sentry.WithScope(func(scope *sentry.Scope) {
		// Add all the tags from the span
		for k, v := range m.tags {
			scope.SetTag(k, v)
		}

		// Add the span context
		scope.SetContext("span", map[string]interface{}{
			"operation": m.operation,
			"name":      m.name,
		})

		// Capture the exception
		sentry.CaptureException(err)
	})

	logger.Errorf("Operation %s (%s) error: %v", m.name, m.operation, err)
}

// MonitorFunc wraps a function with monitoring
func MonitorFunc(ctx context.Context, name, operation string, fn func(context.Context) error) error {
	op, opCtx := StartOperation(ctx, name, operation)
	defer op.Finish()

	err := fn(opCtx)
	if err != nil {
		op.RecordError(err)
	}

	return err
}
