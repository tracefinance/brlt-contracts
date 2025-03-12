// Package logger provides a logging abstraction layer with structured logging capabilities
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.Logger
}

// NewLogger creates a new Logger implementation using Zap
func NewLogger(opts ...Option) (Logger, error) {
	config := zap.NewProductionConfig()

	// Apply options
	for _, opt := range opts {
		if err := opt(&config); err != nil {
			return nil, err
		}
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{logger: logger}, nil
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, convertFields(fields...)...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, convertFields(fields...)...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, convertFields(fields...)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, convertFields(fields...)...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, convertFields(fields...)...)
}

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		logger: l.logger.With(convertFields(fields...)...),
	}
}

// convertFields converts our Field type to zap.Field
func convertFields(fields ...Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(field.Key, field.Value)
	}
	return zapFields
}

// Common options for Zap logger configuration

// WithLevel sets the minimum log level
func WithLevel(level string) Option {
	return func(cfg any) error {
		if c, ok := cfg.(*zap.Config); ok {
			var zapLevel zapcore.Level
			if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
				return err
			}
			c.Level = zap.NewAtomicLevelAt(zapLevel)
		}
		return nil
	}
}

// WithDevelopment sets development mode
func WithDevelopment(enabled bool) Option {
	return func(cfg any) error {
		if c, ok := cfg.(*zap.Config); ok {
			c.Development = enabled
			if enabled {
				c.EncoderConfig = zap.NewDevelopmentEncoderConfig()
			}
		}
		return nil
	}
}

// WithOutputPaths sets the output paths for the logger
func WithOutputPaths(paths ...string) Option {
	return func(cfg any) error {
		if c, ok := cfg.(*zap.Config); ok {
			c.OutputPaths = paths
		}
		return nil
	}
}
