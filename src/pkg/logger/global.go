package logger

// Global logger instance
var globalLogger = DefaultLogger()

// SetGlobalLogger sets the global logger
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// Global logging functions

// Debug logs a debug message
func Debug(v ...interface{}) {
	globalLogger.Debug(v...)
}

// Debugf logs a formatted debug message
func Debugf(format string, v ...interface{}) {
	globalLogger.Debugf(format, v...)
}

// Info logs an info message
func Info(v ...interface{}) {
	globalLogger.Info(v...)
}

// Infof logs a formatted info message
func Infof(format string, v ...interface{}) {
	globalLogger.Infof(format, v...)
}

// Warning logs a warning message
func Warning(v ...interface{}) {
	globalLogger.Warning(v...)
}

// Warningf logs a formatted warning message
func Warningf(format string, v ...interface{}) {
	globalLogger.Warningf(format, v...)
}

// Error logs an error message
func Error(v ...interface{}) {
	globalLogger.Error(v...)
}

// Errorf logs a formatted error message
func Errorf(format string, v ...interface{}) {
	globalLogger.Errorf(format, v...)
}

// Fatal logs a fatal message and exits
func Fatal(v ...interface{}) {
	globalLogger.Fatal(v...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, v ...interface{}) {
	globalLogger.Fatalf(format, v...)
}
