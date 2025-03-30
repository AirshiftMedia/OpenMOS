package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Logger levels
const (
	LevelDebug = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

// LevelNames maps level numbers to string names
var LevelNames = map[int]string{
	LevelDebug:   "debug",
	LevelInfo:    "info",
	LevelWarning: "warning",
	LevelError:   "error",
	LevelFatal:   "fatal",
}

// LevelValues maps level names to values
var LevelValues = map[string]int{
	"debug":   LevelDebug,
	"info":    LevelInfo,
	"warning": LevelWarning,
	"warn":    LevelWarning,
	"error":   LevelError,
	"fatal":   LevelFatal,
}

// Logger represents a simple logger
type Logger struct {
	level  int
	logger *log.Logger
}

// NewLogger creates a new logger with the specified level and output
func NewLogger(level int, out io.Writer) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(out, "", log.LstdFlags),
	}
}

// DefaultLogger returns a default logger writing to stdout
func DefaultLogger() *Logger {
	return NewLogger(LevelInfo, os.Stdout)
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level int) {
	l.level = level
}

// GetLevel returns the current logging level
func (l *Logger) GetLevel() int {
	return l.level
}

// log prints a log message with the specified level and caller information
func (l *Logger) log(level int, levelStr string, v ...interface{}) {
	if level < l.level {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}

	// Format the message
	message := fmt.Sprint(v...)

	// Log with timestamp, level, and caller information
	l.logger.Printf("[%s] %s:%d - %s", levelStr, filepath.Base(file), line, message)
}

// logf prints a formatted log message with the specified level and caller information
func (l *Logger) logf(level int, levelStr string, format string, v ...interface{}) {
	if level < l.level {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}

	// Format the message
	message := fmt.Sprintf(format, v...)

	// Log with timestamp, level, and caller information
	l.logger.Printf("[%s] %s:%d - %s", levelStr, filepath.Base(file), line, message)
}

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
