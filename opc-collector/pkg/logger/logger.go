package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/yourusername/opc-collector/pkg/config"
)

// Log is the global logger instance
var Log *zap.Logger

// Init initializes the global logger based on configuration
func Init(cfg config.LoggingConfig) error {
	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Configure encoder based on format
	var encoderConfig zapcore.EncoderConfig
	if cfg.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Configure time format
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create encoder
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Configure output
	var writer zapcore.WriteSyncer
	if cfg.Output == "stdout" || cfg.Output == "" {
		writer = zapcore.AddSync(os.Stdout)
	} else {
		// File output with rotation
		writer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Output,
			MaxSize:    cfg.MaxSizeMB,    // megabytes
			MaxBackups: cfg.MaxBackups,   // number of backups
			MaxAge:     cfg.MaxAgeDays,   // days
			Compress:   true,             // compress rotated files
		})
	}

	// Create core
	core := zapcore.NewCore(encoder, writer, level)

	// Create logger
	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	if Log != nil {
		return Log.Sync()
	}
	return nil
}

// Named creates a named logger (for sub-components)
func Named(name string) *zap.Logger {
	if Log == nil {
		// Fallback to default logger
		Log, _ = zap.NewProduction()
	}
	return Log.Named(name)
}

// With creates a logger with additional fields
func With(fields ...zap.Field) *zap.Logger {
	if Log == nil {
		// Fallback to default logger
		Log, _ = zap.NewProduction()
	}
	return Log.With(fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Fatal(msg, fields...)
	}
}

// Panic logs a panic message and panics
func Panic(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Panic(msg, fields...)
	}
}
