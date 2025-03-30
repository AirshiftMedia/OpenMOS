package server

import (
	"airshift/openmos/pkg/logger"
	"time"

	"github.com/getsentry/sentry-go"
)

// trackError captures an error with Sentry and returns it
func (c *ClientConnection) trackError(err error, operationType string, details map[string]interface{}) error {
	if err == nil {
		return nil
	}

	// Create tags and context for error tracking
	tags := map[string]string{
		"client_id":      c.id,
		"operation_type": operationType,
	}

	// Add extra context if provided
	if details == nil {
		details = make(map[string]interface{})
	}

	// Add client connection info to context
	details["remote_addr"] = c.conn.RemoteAddr().String()
	if deadline, ok := c.conn.RemoteAddr().(interface{ Deadline() (time.Time, bool) }); ok {
		if d, hasDeadline := deadline.Deadline(); hasDeadline {
			details["deadline"] = d.String()
		}
	}

	// Use a scope to capture error with all context
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTags(tags)
		scope.SetContext("client", details)
		scope.SetLevel(sentry.LevelError)
		sentry.CaptureException(err)
	})

	// Log locally as well
	logger.Errorf("[Client %s] %s error: %v", c.id, operationType, err)

	return err
}
