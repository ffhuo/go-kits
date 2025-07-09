package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/ffhuo/go-kits/logger_v2"
)

func main() {
	// Example 1: Basic usage with default settings
	logger, err := logger_v2.New()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	logger.Info(ctx, "This is an info message")
	logger.Debug(ctx, "This is a debug message (won't show with default level)")
	logger.Warn(ctx, "This is a warning message")
	logger.Error(ctx, "This is an error message")

	// Example 2: Logger with custom options
	logger2, err := logger_v2.New(
		logger_v2.WithDebugLevel(),
		logger_v2.WithFormatJSON(),
		logger_v2.WithSource(),
		logger_v2.WithField("service", "example"),
		logger_v2.WithField("version", "1.0.0"),
	)
	if err != nil {
		panic(err)
	}

	logger2.Debug(ctx, "Debug message with JSON format and source info")
	logger2.Info(ctx, "Info message with custom fields")

	// Example 3: Logger with file output and rotation
	logger3, err := logger_v2.New(
		logger_v2.WithInfoLevel(),
		logger_v2.WithFileRotationP("logs/app.log", 10, 5, 7), // 10MB, 5 backups, 7 days
		logger_v2.WithFormatJSON(),
	)
	if err != nil {
		panic(err)
	}

	logger3.Info(ctx, "This message will be written to file")

	// Example 4: Using with context fields
	ctx = logger_v2.WithFields(ctx,
		logger_v2.NewMeta("request_id", "12345"),
		logger_v2.NewMeta("user_id", "user123"),
	)

	logger.Info(ctx, "Message with context fields")

	// Example 5: Using slog.Attr directly
	logger4 := logger.With(
		slog.String("component", "database"),
		slog.Int("connection_pool", 10),
	)
	logger4.Info(ctx, "Database connection established")

	// Example 6: Using groups
	logger5 := logger.WithGroup("http")
	logger5.Info(ctx, "HTTP request processed",
		slog.String("method", "GET"),
		slog.String("path", "/api/users"),
		slog.Int("status", 200),
		slog.Duration("duration", 45*time.Millisecond),
	)

	// Example 7: Configuration-based logger
	config := logger_v2.Config{
		Level:          logger_v2.InfoLevel,
		Format:         "json",
		FilePath:       "logs/config-based.log",
		MaxSize:        50,
		MaxAge:         30,
		MaxBackups:     10,
		DisableConsole: false,
		TimeLayout:     time.RFC3339,
		AddSource:      true,
		Fields: map[string]string{
			"app":     "example",
			"env":     "development",
			"version": "v1.0.0",
		},
	}

	logger6, err := logger_v2.NewFromConfig(config)
	if err != nil {
		panic(err)
	}

	logger6.Info(ctx, "Logger created from configuration")

	// Example 8: Check log levels
	if logger.DebugEnabled() {
		logger.Debug(ctx, "Debug is enabled")
	}

	if logger.InfoEnabled() {
		logger.Info(ctx, "Info is enabled")
	}

	// Example 9: Structured logging with various types
	logger.Info(ctx, "User action",
		slog.String("action", "login"),
		slog.String("username", "john_doe"),
		slog.Time("timestamp", time.Now()),
		slog.Bool("success", true),
		slog.Int("attempt", 1),
		slog.Float64("duration", 1.23),
		slog.Any("metadata", map[string]interface{}{
			"ip":         "192.168.1.1",
			"user_agent": "Mozilla/5.0...",
		}),
	)

	// Example 10: Error with stack trace context
	err = performOperation()
	if err != nil {
		logger.Error(ctx, "Operation failed",
			slog.Any("error", err),
			slog.String("operation", "performOperation"),
		)
	}
}

func performOperation() error {
	// Simulate an operation that might fail
	return nil
}
