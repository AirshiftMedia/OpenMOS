package logger

import "github.com/getsentry/sentry-go"

// CaptureException sends an error to Sentry
func (l *SentryLogger) CaptureException(err error, tags map[string]string, extras map[string]interface{}) {
	if err == nil {
		return
	}

	// Log the error locally first
	l.Error(err)

	// Add tags
	if len(tags) > 0 {
		sentry.WithScope(func(scope *sentry.Scope) {
			// Add tags
			for k, v := range tags {
				scope.SetTag(k, v)
			}

			// Extras were removed from sentry-go; send them as one context
			if len(extras) > 0 {
				scope.SetContext("extra", sentry.Context(extras))
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
		if len(tags) > 0 {
			for k, v := range tags {
				scope.SetTag(k, v)
			}
		}

		// Extras were removed from sentry-go; send them as one context
		if len(extras) > 0 {
			scope.SetContext("extra", sentry.Context(extras))
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
