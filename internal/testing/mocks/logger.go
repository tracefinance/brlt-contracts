package mocks

import (
	"testing"
	"vault0/internal/logger"
)

// NopLogger is a simple logger that does nothing for testing
type NopLogger struct{}

func (l *NopLogger) Debug(msg string, fields ...logger.Field)  {}
func (l *NopLogger) Info(msg string, fields ...logger.Field)   {}
func (l *NopLogger) Warn(msg string, fields ...logger.Field)   {}
func (l *NopLogger) Error(msg string, fields ...logger.Field)  {}
func (l *NopLogger) Fatal(msg string, fields ...logger.Field)  {}
func (l *NopLogger) With(fields ...logger.Field) logger.Logger { return l }

// NewNopLogger creates a new no-op logger suitable for testing
func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

// DebugLogger implements logger.Logger for tests with output to testing.T
type DebugLogger struct {
	T *testing.T
}

// Debug logs a message at debug level to testing output
func (l *DebugLogger) Debug(msg string, fields ...logger.Field) {
	l.T.Logf("DEBUG: %s %v", msg, fields)
}

// Info logs a message at info level to testing output
func (l *DebugLogger) Info(msg string, fields ...logger.Field) {
	l.T.Logf("INFO: %s %v", msg, fields)
}

// Warn logs a message at warn level to testing output
func (l *DebugLogger) Warn(msg string, fields ...logger.Field) {
	l.T.Logf("WARN: %s %v", msg, fields)
}

// Error logs a message at error level to testing output
func (l *DebugLogger) Error(msg string, fields ...logger.Field) {
	l.T.Logf("ERROR: %s %v", msg, fields)
}

// Fatal logs a message at fatal level to testing output
func (l *DebugLogger) Fatal(msg string, fields ...logger.Field) {
	l.T.Logf("FATAL: %s %v", msg, fields)
}

// With returns the same logger (chainable method)
func (l *DebugLogger) With(fields ...logger.Field) logger.Logger {
	return l
}

// NewDebugLogger creates a new debug logger that outputs to testing.T
func NewDebugLogger(t *testing.T) *DebugLogger {
	return &DebugLogger{T: t}
}
