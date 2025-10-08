package logger

import (
	"fmt"

	"github.com/tanint/go-eda/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Initialize creates a new logger based on the configuration
func Initialize(cfg config.LoggerConfig) error {
	var zapCfg zap.Config

	if cfg.Encoding == "console" {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapCfg = zap.NewProductionConfig()
	}

	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	// Set output path
	if cfg.OutputPath != "" {
		zapCfg.OutputPaths = []string{cfg.OutputPath}
	}

	logger, err := zapCfg.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	log = logger
	return nil
}

// Get returns the global logger instance
func Get() *zap.Logger {
	if log == nil {
		// Fallback to a default logger if not initialized
		log, _ = zap.NewProduction()
	}
	return log
}

// Sync flushes any buffered log entries
func Sync() error {
	if log != nil {
		return log.Sync()
	}
	return nil
}

// With creates a child logger with additional fields
func With(fields ...zap.Field) *zap.Logger {
	return Get().With(fields...)
}

// Info logs an info level message
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Error logs an error level message
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Debug logs a debug level message
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Warn logs a warning level message
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Fatal logs a fatal level message and exits
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}
