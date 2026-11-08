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
