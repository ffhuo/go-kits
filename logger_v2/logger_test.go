package logger_v2

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	logger, err := New()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	if logger == nil {
		t.Fatal("Logger is nil")
	}
}

func TestWithOptions(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(
		WithDebugLevel(),
		WithFormatJSON(),
		WithField("test", "value"),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Redirect output to buffer for testing
	logger.logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx := context.Background()
	logger.Debug(ctx, "debug message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("Expected debug message in output, got: %s", output)
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(WithDebugLevel())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Redirect output to buffer
	logger.logger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx := context.Background()
	logger.Debug(ctx, "debug message")
	logger.Info(ctx, "info message")
	logger.Warn(ctx, "warn message")
	logger.Error(ctx, "error message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Error("Debug message not found")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info message not found")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message not found")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message not found")
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()
	ctx = WithFields(ctx,
		NewMeta("request_id", "12345"),
		NewMeta("user_id", "user123"),
	)

	logger.Info(ctx, "test message")

	output := buf.String()
	if !strings.Contains(output, "request_id") {
		t.Error("request_id not found in output")
	}
	if !strings.Contains(output, "12345") {
		t.Error("request_id value not found in output")
	}
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	childLogger := logger.With(
		slog.String("component", "test"),
		slog.Int("version", 1),
	)

	ctx := context.Background()
	childLogger.Info(ctx, "test message")

	output := buf.String()
	if !strings.Contains(output, "component") {
		t.Error("component field not found")
	}
	if !strings.Contains(output, "test") {
		t.Error("component value not found")
	}
}

func TestWithGroup(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	groupLogger := logger.WithGroup("http")
	ctx := context.Background()
	groupLogger.logger.InfoContext(ctx, "request processed",
		slog.String("method", "GET"),
		slog.Int("status", 200),
	)

	output := buf.String()
	if !strings.Contains(output, "http") {
		t.Error("group name not found")
	}
}

func TestLevelEnabled(t *testing.T) {
	logger, err := New(WithInfoLevel())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	if logger.DebugEnabled() {
		t.Error("Debug should not be enabled with Info level")
	}
	if !logger.InfoEnabled() {
		t.Error("Info should be enabled with Info level")
	}
	if !logger.WarnEnabled() {
		t.Error("Warn should be enabled with Info level")
	}
	if !logger.ErrorEnabled() {
		t.Error("Error should be enabled with Info level")
	}
}

func TestNewFromConfig(t *testing.T) {
	config := Config{
		Level:      InfoLevel,
		Format:     "json",
		AddSource:  true,
		TimeLayout: time.RFC3339,
		Fields: map[string]string{
			"app": "test",
			"env": "testing",
		},
	}

	logger, err := NewFromConfig(config)
	if err != nil {
		t.Fatalf("Failed to create logger from config: %v", err)
	}
	if logger == nil {
		t.Fatal("Logger is nil")
	}
}

func TestTrace(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(WithDebugLevel())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx := context.Background()
	begin := time.Now()

	logger.Trace(ctx, begin, func() (string, int64) {
		return "SELECT * FROM users", 5
	}, nil)

	output := buf.String()
	if !strings.Contains(output, "SELECT * FROM users") {
		t.Error("SQL query not found in trace output")
	}
}

func TestFileOutput(t *testing.T) {
	// Create a temporary file-like buffer
	var buf bytes.Buffer

	// Create logger with custom writer (simulating file output)
	logger := &Logger{
		logger: slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}

	ctx := context.Background()
	logger.Info(ctx, "test file output")

	output := buf.String()
	if !strings.Contains(output, "test file output") {
		t.Error("Message not found in file output")
	}
}

func TestMeta(t *testing.T) {
	meta := NewMeta("test_key", "test_value")
	if meta.Key() != "test_key" {
		t.Errorf("Expected key 'test_key', got '%s'", meta.Key())
	}
	if meta.Value() != "test_value" {
		t.Errorf("Expected value 'test_value', got '%v'", meta.Value())
	}
}

func TestSlogAttrs(t *testing.T) {
	metas := []Meta{
		NewMeta("key1", "value1"),
		NewMeta("key2", 42),
	}

	attrs := SlogAttrs(metas)
	if len(attrs) != 2 {
		t.Errorf("Expected 2 attributes, got %d", len(attrs))
	}
}

func TestSlogAny(t *testing.T) {
	metas := []Meta{
		NewMeta("key1", "value1"),
		NewMeta("key2", 42),
	}

	fields := SlogAny(metas)
	if len(fields) != 4 { // 2 metas * 2 (key-value pairs)
		t.Errorf("Expected 4 fields, got %d", len(fields))
	}
}

// Benchmark tests
func BenchmarkLogger_Info(b *testing.B) {
	logger, _ := New(WithInfoLevel())
	logger.logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info(ctx, "benchmark message")
	}
}

func BenchmarkLogger_InfoWithFields(b *testing.B) {
	logger, _ := New(WithInfoLevel())
	logger.logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()
	ctx = WithFields(ctx,
		NewMeta("request_id", "12345"),
		NewMeta("user_id", "user123"),
	)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info(ctx, "benchmark message")
	}
}

func BenchmarkLogger_InfoStructured(b *testing.B) {
	logger, _ := New(WithInfoLevel())
	logger.logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.logger.InfoContext(ctx, "benchmark message",
			slog.String("key1", "value1"),
			slog.Int("key2", 42),
		)
	}
}
