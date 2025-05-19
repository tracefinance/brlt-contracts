package mocks

import (
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
