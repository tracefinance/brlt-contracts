// Package logger provides a logging abstraction layer with structured logging capabilities
package logger

import "time"

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value any
}

// Logger is the interface that wraps the basic logging methods
type Logger interface {
	// Debug logs a message at debug level
	Debug(msg string, fields ...Field)
	// Info logs a message at info level
	Info(msg string, fields ...Field)
	// Warn logs a message at warn level
	Warn(msg string, fields ...Field)
	// Error logs a message at error level
	Error(msg string, fields ...Field)
	// Fatal logs a message at fatal level and then calls os.Exit(1)
	Fatal(msg string, fields ...Field)
	// With returns a new Logger with the given fields added to it
	With(fields ...Field) Logger
}

// Option represents a configuration option for the logger
type Option func(any) error

// LoggerFactory creates a new logger instance with the given options
type LoggerFactory func(opts ...Option) (Logger, error)

// String creates a Field with a string value
func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates a Field with an int value
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates a Field with an int64 value
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a Field with a float64 value
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a Field with a bool value
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a Field with a time.Duration value
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Time creates a Field with a time.Time value
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Error creates a Field with an error value
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

// Any creates a Field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
