// Package logger provides a logging abstraction layer with structured logging capabilities
package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"vault0/internal/config"
)

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

// zapLogger implements Logger interface using zap
type zapLogger struct {
	logger *zap.Logger
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(cfg config.LogConfig) (Logger, error) {
	// Create encoder config based on format
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Configure encoder based on format
	var encoder zapcore.Encoder
	if cfg.Format == config.LogFormatJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Configure log level
	var level zapcore.Level
	switch cfg.Level {
	case config.LogLevelDebug:
		level = zapcore.DebugLevel
	case config.LogLevelInfo:
		level = zapcore.InfoLevel
	case config.LogLevelWarn:
		level = zapcore.WarnLevel
	case config.LogLevelError:
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Configure output
	var output zapcore.WriteSyncer
	if cfg.OutputPath != "" {
		file, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = zapcore.AddSync(file)
	} else {
		output = zapcore.AddSync(os.Stdout)
	}

	// Create core
	core := zapcore.NewCore(encoder, output, level)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &zapLogger{logger: logger}, nil
}

// Debug implements Logger.Debug
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fieldsToZapFields(fields)...)
}

// Info implements Logger.Info
func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fieldsToZapFields(fields)...)
}

// Warn implements Logger.Warn
func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fieldsToZapFields(fields)...)
}

// Error implements Logger.Error
func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fieldsToZapFields(fields)...)
}

// Fatal implements Logger.Fatal
func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, fieldsToZapFields(fields)...)
}

// With implements Logger.With
func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		logger: l.logger.With(fieldsToZapFields(fields)...),
	}
}

// fieldsToZapFields converts our Field type to zap.Field
func fieldsToZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = toZapField(f)
	}
	return zapFields
}

// toZapField converts a single Field to zap.Field
func toZapField(f Field) zap.Field {
	switch v := f.Value.(type) {
	case string:
		return zap.String(f.Key, v)
	case int:
		return zap.Int(f.Key, v)
	case int64:
		return zap.Int64(f.Key, v)
	case float64:
		return zap.Float64(f.Key, v)
	case bool:
		return zap.Bool(f.Key, v)
	case time.Duration:
		return zap.Duration(f.Key, v)
	case time.Time:
		return zap.Time(f.Key, v)
	case error:
		return zap.Error(v)
	default:
		return zap.Any(f.Key, v)
	}
}

// Field constructors
func String(key string, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
