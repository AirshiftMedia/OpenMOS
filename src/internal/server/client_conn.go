package server

import (
	"airshift/openmos/pkg/logger"
	"fmt"
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

// Write sends data to the client
func (c *ClientConnection) Write(data []byte) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	// Set write deadline
	err := c.conn.SetWriteDeadline(time.Now().Add(c.config.Server.WriteTimeout))
	if err != nil {
		return c.trackError(err, "set_write_deadline", nil)
	}

	_, err = c.conn.Write(data)
	if err != nil {
		return c.trackError(err, "write", map[string]interface{}{
			"data_length": len(data),
		})
	}

	return nil
}

// Close closes the client connection
func (c *ClientConnection) Close() {
	c.closeOnce.Do(func() {
		logger.Infof("Closing connection for client %s", c.id)
		close(c.closeChan)

		// Track connection closure in Sentry
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("client_id", c.id)
			scope.SetLevel(sentry.LevelInfo)
			sentry.CaptureMessage(fmt.Sprintf("Client connection %s closed", c.id))
		})

		if err := c.conn.Close(); err != nil {
			c.trackError(err, "close", nil)
		}

		c.server.unregisterClient(c.id)
	})
}

// ID returns the client ID
func (c *ClientConnection) ID() string {
	return c.id
}
