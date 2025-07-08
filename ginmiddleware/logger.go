package ginmiddleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
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
	Logger          Logger        // 日志实例
	SkipPaths       []string      // 跳过记录的路径
	SlowThreshold   time.Duration // 慢请求阈值
	ShowQueryParams bool          // 是否显示GET请求的query参数
	ShowRequestBody bool          // 是否显示POST/PUT/PATCH请求的body内容
	MaxBodySize     int           // 最大body显示大小（字节）
	SkipBodyMethods []string      // 跳过显示body的HTTP方法
	SensitiveFields []string      // 敏感字段，会被脱敏显示
}

// DefaultLoggerConfig 默认日志配置
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		SkipPaths:       []string{"/health", "/metrics"},
		SlowThreshold:   200 * time.Millisecond,
		ShowQueryParams: true,
		ShowRequestBody: true,
		MaxBodySize:     1024, // 1KB
		SkipBodyMethods: []string{"GET", "HEAD", "OPTIONS"},
		SensitiveFields: []string{"password", "token", "secret", "key"},
	}
}

// bodyLogWriter 用于捕获响应内容的writer
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
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

		// 获取请求参数
		var requestParams string
		method := c.Request.Method

		// 处理GET请求的query参数
		if method == "GET" && cfg.ShowQueryParams && c.Request.URL.RawQuery != "" {
			requestParams = fmt.Sprintf("Query: %s", maskSensitiveData(c.Request.URL.RawQuery, cfg.SensitiveFields))
		}

		// 处理POST/PUT/PATCH请求的body内容
		if cfg.ShowRequestBody && !containsIgnoreCase(cfg.SkipBodyMethods, method) {
			if body := readRequestBody(c, cfg.MaxBodySize); body != "" {
				requestParams = fmt.Sprintf("Body: %s", maskSensitiveData(body, cfg.SensitiveFields))
			}
		}

		// 处理请求
		c.Next()

		// 计算响应时间
		latency := time.Since(start)

		// 构建日志信息
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 构建基础日志消息
		logParts := []string{
			fmt.Sprintf("RequestID: %s", requestID),
			fmt.Sprintf("Method: %s", method),
			fmt.Sprintf("Path: %s", path),
			fmt.Sprintf("Status: %d", statusCode),
			fmt.Sprintf("ClientIP: %s", clientIP),
			fmt.Sprintf("Latency: %v", latency),
		}

		// 添加请求参数信息
		if requestParams != "" {
			logParts = append(logParts, requestParams)
		}

		// 添加User-Agent（截取前100个字符）
		if userAgent != "" {
			if len(userAgent) > 100 {
				userAgent = userAgent[:100] + "..."
			}
			logParts = append(logParts, fmt.Sprintf("UserAgent: %s", userAgent))
		}

		// 添加错误信息
		if len(c.Errors) > 0 {
			logParts = append(logParts, fmt.Sprintf("Errors: %s", c.Errors.String()))
		}

		logMsg := strings.Join(logParts, " | ")

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

// readRequestBody 读取请求体内容
func readRequestBody(c *gin.Context, maxSize int) string {
	if c.Request.Body == nil {
		return ""
	}

	// 读取body内容
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, int64(maxSize)))
	if err != nil {
		return ""
	}

	// 恢复body，以便后续处理器可以读取
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 如果内容为空，返回空字符串
	if len(bodyBytes) == 0 {
		return ""
	}

	// 检查是否是文本内容
	body := string(bodyBytes)
	if !isPrintableText(body) {
		return fmt.Sprintf("[Binary data, %d bytes]", len(bodyBytes))
	}

	// 如果超过最大长度，截断并添加省略号
	if len(bodyBytes) >= maxSize {
		body += "..."
	}

	return body
}

// SimpleLoggerMiddleware 简化版日志中间件，只记录基本信息
func SimpleLoggerMiddleware(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := GetRequestID(c)

		// 处理请求
		c.Next()

		// 记录请求信息
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Info(c.Request.Context(),
			"HTTP Request",
			fmt.Sprintf("RequestID: %s | Method: %s | Path: %s | Status: %d | Latency: %v | ClientIP: %s",
				requestID,
				c.Request.Method,
				c.Request.URL.Path,
				statusCode,
				latency,
				c.ClientIP(),
			),
		)
	}
}
