package logger

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	gorm "gorm.io/gorm/logger"
)

func TestLogger(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")

	// 创建日志实例
	logger, err := New(
		WithDebugLevel(),
		WithFormatJSON(),
		WithFileRotationP(logFile, 1, 3, 1),
		WithField("app", "test"),
		WithField("env", "dev"),
		WithTimeLayout(time.RFC3339),
	)
	if err != nil {
		t.Fatal(err)
	}

	// 测试基本日志
	ctx := context.Background()
	logger.Debug(ctx, "debug message: %s", "test")
	logger.Info(ctx, "info message: %s", "test")
	logger.Warn(ctx, "warn message: %s", "test")
	logger.Error(ctx, "error message: %s", "test")

	// 测试带 Fields 的日志
	ctx = WithFields(ctx, NewMeta("requestID", "123456"))
	logger.Info(ctx, "request received")

	// 测试 Gin Context
	ginCtx := &gin.Context{}
	ginCtx.Set(MetaKey, []Meta{NewMeta("traceID", "abcdef")})
	logger.Info(ginCtx, "gin request received")

	// 测试多个 Fields
	ctx = WithFields(ctx,
		NewMeta("userID", "user123"),
		NewMeta("action", "login"),
	)
	logger.Info(ctx, "user action")

	// 测试 Trace
	begin := time.Now()
	logger.Trace(ctx, begin, func() (string, int64) {
		return "SELECT * FROM users WHERE id = ?", 1
	}, nil)

	// 测试 Trace 带错误
	logger.Trace(ctx, begin, func() (string, int64) {
		return "INSERT INTO users (name) VALUES (?)", 0
	}, gorm.ErrRecordNotFound)

	// 验证日志文件是否创建
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("log file was not created")
	}

	// 读取日志文件内容
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	// 验证日志内容
	if len(content) == 0 {
		t.Error("log file is empty")
	}
}

func TestLoggerWithoutConsole(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")

	// 创建日志实例
	logger, err := New(
		WithInfoLevel(),
		WithFileP(logFile),
		WithDisableConsole(),
	)
	if err != nil {
		t.Fatal(err)
	}

	// 测试基本日志
	ctx := context.Background()
	logger.Info(ctx, "test message")

	// 测试带 Fields 的日志
	ctx = WithFields(ctx, NewMeta("requestID", "123456"))
	logger.Info(ctx, "request received")

	// 验证日志文件
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) == 0 {
		t.Error("log file is empty")
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name          string
		level         Option
		shouldLogInfo bool
	}{
		{
			name:          "debug level",
			level:         WithDebugLevel(),
			shouldLogInfo: true,
		},
		{
			name:          "info level",
			level:         WithInfoLevel(),
			shouldLogInfo: true,
		},
		{
			name:          "warn level",
			level:         WithWarnLevel(),
			shouldLogInfo: false,
		},
		{
			name:          "error level",
			level:         WithErrorLevel(),
			shouldLogInfo: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建临时目录
			tmpDir, err := os.MkdirTemp("", "logger_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			logFile := filepath.Join(tmpDir, "test.log")

			// 创建日志实例
			logger, err := New(
				tt.level,
				WithFileP(logFile),
				WithDisableConsole(),
			)
			if err != nil {
				t.Fatal(err)
			}

			// 测试基本日志
			ctx := context.Background()
			logger.Info(ctx, "test info message")

			// 测试带 Fields 的日志
			ctx = WithFields(ctx, NewMeta("level", tt.name))
			logger.Info(ctx, "test info message with fields")

			// 检查日志文件内容
			content, err := os.ReadFile(logFile)
			if err != nil {
				t.Fatal(err)
			}

			hasContent := len(content) > 0
			if hasContent != tt.shouldLogInfo {
				t.Errorf("expected shouldLogInfo=%v, got log content: %v", tt.shouldLogInfo, hasContent)
			}
		})
	}
}

func TestLoggerFields(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")

	// 创建日志实例
	logger, err := New(
		WithDebugLevel(),
		WithFormatJSON(),
		WithFileP(logFile),
		WithDisableConsole(),
	)
	if err != nil {
		t.Fatal(err)
	}

	// 测试不同类型的 Fields
	tests := []struct {
		name   string
		fields []Meta
	}{
		{
			name: "string fields",
			fields: []Meta{
				NewMeta("requestID", "123456"),
				NewMeta("userID", "user123"),
			},
		},
		{
			name: "mixed type fields",
			fields: []Meta{
				NewMeta("statusCode", 200),
				NewMeta("success", true),
				NewMeta("duration", 123.456),
			},
		},
		{
			name: "nested fields",
			fields: []Meta{
				NewMeta("user", map[string]interface{}{
					"id":   "user123",
					"name": "John Doe",
					"role": "admin",
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = WithFields(ctx, tt.fields...)
			logger.Info(ctx, "test message with fields")
		})
	}

	// 验证日志文件
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) == 0 {
		t.Error("log file is empty")
	}
}
