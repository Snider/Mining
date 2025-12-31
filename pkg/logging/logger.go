// Package logging provides structured logging with log levels and fields.
package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	// LevelDebug is the most verbose log level.
	LevelDebug Level = iota
	// LevelInfo is for general informational messages.
	LevelInfo
	// LevelWarn is for warning messages.
	LevelWarn
	// LevelError is for error messages.
	LevelError
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging with configurable output and level.
type Logger struct {
	mu        sync.Mutex
	output    io.Writer
	level     Level
	component string
}

// Config holds configuration for creating a new Logger.
type Config struct {
	Output    io.Writer
	Level     Level
	Component string
}

// DefaultConfig returns the default logger configuration.
func DefaultConfig() Config {
	return Config{
		Output:    os.Stderr,
		Level:     LevelInfo,
		Component: "",
	}
}

// New creates a new Logger with the given configuration.
func New(cfg Config) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}
	return &Logger{
		output:    cfg.Output,
		level:     cfg.Level,
		component: cfg.Component,
	}
}

// WithComponent returns a new Logger with the specified component name.
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		output:    l.output,
		level:     l.level,
		component: component,
	}
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level.
func (l *Logger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// Fields represents key-value pairs for structured logging.
type Fields map[string]interface{}

// log writes a log message at the specified level.
func (l *Logger) log(level Level, msg string, fields Fields) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	// Build the log line
	var sb strings.Builder
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	sb.WriteString(timestamp)
	sb.WriteString(" [")
	sb.WriteString(level.String())
	sb.WriteString("]")

	if l.component != "" {
		sb.WriteString(" [")
		sb.WriteString(l.component)
		sb.WriteString("]")
	}

	sb.WriteString(" ")
	sb.WriteString(msg)

	// Add fields if present
	if len(fields) > 0 {
		sb.WriteString(" |")
		for k, v := range fields {
			sb.WriteString(" ")
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(fmt.Sprintf("%v", v))
		}
	}

	sb.WriteString("\n")
	fmt.Fprint(l.output, sb.String())
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields ...Fields) {
	l.log(LevelDebug, msg, mergeFields(fields))
}

// Info logs an informational message.
func (l *Logger) Info(msg string, fields ...Fields) {
	l.log(LevelInfo, msg, mergeFields(fields))
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields ...Fields) {
	l.log(LevelWarn, msg, mergeFields(fields))
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields ...Fields) {
	l.log(LevelError, msg, mergeFields(fields))
}

// Debugf logs a formatted debug message.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, fmt.Sprintf(format, args...), nil)
}

// Infof logs a formatted informational message.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, fmt.Sprintf(format, args...), nil)
}

// Warnf logs a formatted warning message.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...), nil)
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...), nil)
}

// mergeFields combines multiple Fields maps into one.
func mergeFields(fields []Fields) Fields {
	if len(fields) == 0 {
		return nil
	}
	result := make(Fields)
	for _, f := range fields {
		for k, v := range f {
			result[k] = v
		}
	}
	return result
}

// --- Global logger for convenience ---

var (
	globalLogger = New(DefaultConfig())
	globalMu     sync.RWMutex
)

// SetGlobal sets the global logger instance.
func SetGlobal(l *Logger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger = l
}

// GetGlobal returns the global logger instance.
func GetGlobal() *Logger {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalLogger
}

// SetGlobalLevel sets the log level of the global logger.
func SetGlobalLevel(level Level) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.SetLevel(level)
}

// Global convenience functions that use the global logger

// Debug logs a debug message using the global logger.
func Debug(msg string, fields ...Fields) {
	GetGlobal().Debug(msg, fields...)
}

// Info logs an informational message using the global logger.
func Info(msg string, fields ...Fields) {
	GetGlobal().Info(msg, fields...)
}

// Warn logs a warning message using the global logger.
func Warn(msg string, fields ...Fields) {
	GetGlobal().Warn(msg, fields...)
}

// Error logs an error message using the global logger.
func Error(msg string, fields ...Fields) {
	GetGlobal().Error(msg, fields...)
}

// Debugf logs a formatted debug message using the global logger.
func Debugf(format string, args ...interface{}) {
	GetGlobal().Debugf(format, args...)
}

// Infof logs a formatted informational message using the global logger.
func Infof(format string, args ...interface{}) {
	GetGlobal().Infof(format, args...)
}

// Warnf logs a formatted warning message using the global logger.
func Warnf(format string, args ...interface{}) {
	GetGlobal().Warnf(format, args...)
}

// Errorf logs a formatted error message using the global logger.
func Errorf(format string, args ...interface{}) {
	GetGlobal().Errorf(format, args...)
}

// ParseLevel parses a string into a log level.
func ParseLevel(s string) (Level, error) {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARN", "WARNING":
		return LevelWarn, nil
	case "ERROR":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level: %s", s)
	}
}
