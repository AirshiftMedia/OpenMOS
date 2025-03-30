package monitoring

import (
	"context"
	"time"

	"airshift/openmos/pkg/logger"

	"github.com/getsentry/sentry-go"
)

// MonitoredOperation represents an operation being monitored
type MonitoredOperation struct {
	transaction *sentry.Transaction
	span        *sentry.Span
	name        string
	operation   string
	start       time.Time
	tags        map[string]string
}

// StartOperation starts monitoring an operation
func StartOperation(ctx context.Context, name, operation string) (*MonitoredOperation, context.Context) {
	// Create transaction or span
	var tx *sentry.Transaction
	var span *sentry.Span
	var opCtx context.Context

	// Try to get parent transaction from context
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		// Check if we have an active transaction already
		if span := sentry.SpanFromContext(ctx); span != nil {
			// Create child span
			childSpan := span.StartChild(operation)
			childSpan.Description = name

			opCtx = sentry.SetSpanContext(ctx, childSpan)
			return &MonitoredOperation{
				transaction: nil,
				span:        childSpan,
				name:        name,
				operation:   operation,
				start:       time.Now(),
				tags:        make(map[string]string),
			}, opCtx
		}
	}

	// No parent transaction, create a new one
	tx = sentry.StartTransaction(sentry.Context{
		Name: name,
		Op:   operation,
	})

	opCtx = sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())
	opCtx = sentry.SetSpanContext(opCtx, tx)

	return &MonitoredOperation{
		transaction: tx,
		span:        nil,
		name:        name,
		operation:   operation,
		start:       time.Now(),
		tags:        make(map[string]string),
	}, opCtx
}

// SetTag adds a tag to the operation
func (m *MonitoredOperation) SetTag(key, value string) {
	if m == nil {
		return
	}

	m.tags[key] = value

	if m.transaction != nil {
		m.transaction.SetTag(key, value)
	} else if m.span != nil {
		m.span.SetTag(key, value)
	}
}

// SetData adds data to the operation
func (m *MonitoredOperation) SetData(key string, value interface{}) {
	if m == nil {
		return
	}

	if m.transaction != nil {
		m.transaction.SetData(key, value)
	} else if m.span != nil {
		m.span.SetData(key, value)
	}
}

// Finish completes the operation
func (m *MonitoredOperation) Finish() {
	if m == nil {
		return
	}

	duration := time.Since(m.start)

	// Complete the operation
	if m.transaction != nil {
		m.transaction.Finish()
	} else if m.span != nil {
		m.span.Finish()
	}

	// Log the operation for local visibility
	logger.Debugf("Operation %s (%s) completed in %v", m.name, m.operation, duration)
}

// RecordError records an error that occurred during the operation
func (m *MonitoredOperation) RecordError(err error) {
	if m == nil || err == nil {
		return
	}

	if m.transaction != nil {
		m.transaction.Status = "internal_error"

		sentry.WithScope(func(scope *sentry.Scope) {
			// Add all the tags from the transaction
			for k, v := range m.tags {
				scope.SetTag(k, v)
			}

			// Set the transaction on the scope
			scope.SetTransaction(m.transaction.Name)

			// Capture the exception
			sentry.CaptureException(err)
		})

	} else if m.span != nil {
		m.span.Status = "internal_error"

		// Also set status on parent transaction if available
		if parent := m.span.TraceID; parent != "" {
			if hub := sentry.CurrentHub(); hub != nil {
				if tx := hub.Scope().Transaction(); tx != nil {
					tx.Status = "internal_error"
				}
			}
		}

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
	}

	logger.Errorf("Operation %s (%s) error: %v", m.name, m.operation, err)
}

// Middleware functions for common operations

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
