package ginmiddleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 通用日志接口，兼容logger和logger_v2
type Logger interface {
	Info(ctx context.Context, msg string, data ...interface{})
	Warn(ctx context.Context, msg string, data ...interface{})
	Error(ctx context.Context, msg string, data ...interface{})
	Debug(ctx context.Context, msg string, data ...interface{})
}

// LoggerConfig 日志中间件配置
type LoggerConfig struct {
	Logger        Logger        // 日志实例
	SkipPaths     []string      // 跳过记录的路径
	TimeFormat    string        // 时间格式
	UTCTime       bool          // 是否使用UTC时间
	SlowThreshold time.Duration // 慢请求阈值
	EnableColors  bool          // 是否启用颜色输出
	CustomFields  []string      // 自定义字段
}

// DefaultLoggerConfig 默认日志配置
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		SkipPaths:     []string{"/health", "/metrics"},
		TimeFormat:    "2006-01-02 15:04:05",
		UTCTime:       false,
		SlowThreshold: 200 * time.Millisecond,
		EnableColors:  false,
		CustomFields:  []string{},
	}
}

// LoggerMiddleware 请求日志中间件
func LoggerMiddleware(config ...*LoggerConfig) gin.HandlerFunc {
	var cfg *LoggerConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = DefaultLoggerConfig()
	}

	if cfg.Logger == nil {
		panic("Logger is required")
	}

	return func(c *gin.Context) {
		// 检查是否跳过记录
		path := c.Request.URL.Path
		for _, skipPath := range cfg.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// 记录开始时间
		start := time.Now()

		// 获取请求ID
		requestID := GetRequestID(c)
		if requestID == "" {
			requestID = "unknown"
		}

		// 处理请求
		c.Next()

		// 计算响应时间
		latency := time.Since(start)

		// 获取时间戳
		timestamp := start
		if cfg.UTCTime {
			timestamp = timestamp.UTC()
		}

		// 构建日志消息
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 获取请求体大小
		bodySize := c.Writer.Size()
		if bodySize < 0 {
			bodySize = 0
		}

		// 构建详细的日志信息
		logMsg := fmt.Sprintf("[%s] %s %s - %d - %s - %s - %s - %dB - %s",
			requestID,
			method,
			path,
			statusCode,
			clientIP,
			userAgent,
			latency,
			bodySize,
			timestamp.Format(cfg.TimeFormat),
		)

		// 添加查询参数
		if c.Request.URL.RawQuery != "" {
			logMsg += fmt.Sprintf(" - Query: %s", c.Request.URL.RawQuery)
		}

		// 添加错误信息
		if len(c.Errors) > 0 {
			logMsg += fmt.Sprintf(" - Errors: %s", c.Errors.String())
		}

		// 根据状态码和响应时间选择日志级别
		ctx := context.Background()

		switch {
		case statusCode >= 500:
			cfg.Logger.Error(ctx, "HTTP Request Error", logMsg)
		case statusCode >= 400:
			cfg.Logger.Warn(ctx, "HTTP Request Warning", logMsg)
		case latency > cfg.SlowThreshold:
			cfg.Logger.Warn(ctx, "HTTP Request Slow", logMsg)
		default:
			cfg.Logger.Info(ctx, "HTTP Request", logMsg)
		}
	}
}

// RequestInfoMiddleware 请求信息记录中间件（更详细的版本）
func RequestInfoMiddleware(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := GetRequestID(c)

		// 记录请求开始
		logger.Info(c.Request.Context(),
			"Request started - ID: %s, Method: %s, Path: %s, IP: %s, UserAgent: %s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		// 处理请求
		c.Next()

		// 记录请求结束
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		logger.Info(c.Request.Context(),
			"Request completed - ID: %s, Status: %d, Duration: %v, Size: %d bytes",
			requestID,
			statusCode,
			latency,
			bodySize,
		)
	}
}
