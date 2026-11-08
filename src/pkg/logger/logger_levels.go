package logger

import "os"

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {
	l.log(LevelDebug, "DEBUG", v...)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logf(LevelDebug, "DEBUG", format, v...)
}

// Info logs an info message
func (l *Logger) Info(v ...interface{}) {
	l.log(LevelInfo, "INFO", v...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.logf(LevelInfo, "INFO", format, v...)
}

// Warning logs a warning message
func (l *Logger) Warning(v ...interface{}) {
	l.log(LevelWarning, "WARN", v...)
}

// Warningf logs a formatted warning message
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.logf(LevelWarning, "WARN", format, v...)
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	l.log(LevelError, "ERROR", v...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logf(LevelError, "ERROR", format, v...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.log(LevelFatal, "FATAL", v...)
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logf(LevelFatal, "FATAL", format, v...)
	os.Exit(1)
}
