package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ffhuo/go-kits/logger"
)

func main() {
	// 创建日志实例
	log, err := logger.New(
		logger.WithDebugLevel(),
		logger.WithFormatJSON(),
		logger.WithFileRotationP("logs/app.log", 100, 7, 10), // 100MB, 7天, 10个备份
		logger.WithField("app", "example"),
		logger.WithField("env", "dev"),
		logger.WithTimeLayout(time.RFC3339),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// 基本日志
	log.Debug(ctx, "Debug message: %s", "hello")
	log.Info(ctx, "Info message: %s", "hello")
	log.Warn(ctx, "Warning message: %s", "hello")
	log.Error(ctx, "Error message: %s", "hello")

	// 带字段的日志
	ctx = logger.WithFields(ctx,
		logger.NewMeta("requestID", "123456"),
		logger.NewMeta("userID", "user123"),
	)
	log.Info(ctx, "Request received")

	// 嵌套字段
	ctx = logger.WithFields(ctx,
		logger.NewMeta("user", map[string]interface{}{
			"id":   "user123",
			"name": "John Doe",
			"role": "admin",
		}),
	)
	log.Info(ctx, "User info")

	// SQL 追踪
	begin := time.Now()
	log.Trace(ctx, begin, func() (string, int64) {
		time.Sleep(200 * time.Millisecond) // 模拟慢查询
		return "SELECT * FROM users WHERE id = ?", 1
	}, nil)

	// 带错误的 SQL 追踪
	log.Trace(ctx, begin, func() (string, int64) {
		return "INSERT INTO users (name) VALUES (?)", 0
	}, fmt.Errorf("failed to insert user: %s", "John Doe"))
}
