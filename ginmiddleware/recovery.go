package ginmiddleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RecoveryConfig Recovery中间件配置
type RecoveryConfig struct {
	Logger           Logger                          // 日志实例
	EnableStackTrace bool                            // 是否启用堆栈跟踪
	StackTraceSize   int                             // 堆栈跟踪大小
	CustomHandler    func(*gin.Context, interface{}) // 自定义错误处理器
	PrintStack       bool                            // 是否打印堆栈到控制台
}

// DefaultRecoveryConfig 默认Recovery配置
func DefaultRecoveryConfig() *RecoveryConfig {
	return &RecoveryConfig{
		EnableStackTrace: true,
		StackTraceSize:   4 << 10, // 4KB
		PrintStack:       true,
	}
}

// Recovery 恢复中间件，处理panic并记录详细错误信息
func Recovery(config ...*RecoveryConfig) gin.HandlerFunc {
	var cfg *RecoveryConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = DefaultRecoveryConfig()
	}

	if cfg.Logger == nil {
		panic("Logger is required for Recovery middleware")
	}

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取请求ID
				requestID := GetRequestID(c)
				if requestID == "" {
					requestID = "unknown"
				}

				// 获取堆栈信息
				var stack []byte
				if cfg.EnableStackTrace {
					stack = debug.Stack()
				}

				// 构建错误信息
				errorMsg := fmt.Sprintf("Panic recovered: %v", err)

				// 构建详细的请求信息
				requestInfo := buildRequestInfo(c, requestID)

				// 记录错误日志
				logPanicError(cfg.Logger, c.Request.Context(), errorMsg, requestInfo, stack, cfg.PrintStack)

				// 如果有自定义处理器，使用自定义处理器
				if cfg.CustomHandler != nil {
					cfg.CustomHandler(c, err)
					return
				}

				// 默认错误响应
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal Server Error",
					"request_id": requestID,
					"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}

// buildRequestInfo 构建请求信息
func buildRequestInfo(c *gin.Context, requestID string) string {
	var info strings.Builder

	info.WriteString(fmt.Sprintf("Request ID: %s\n", requestID))
	info.WriteString(fmt.Sprintf("Method: %s\n", c.Request.Method))
	info.WriteString(fmt.Sprintf("Path: %s\n", c.Request.URL.Path))
	info.WriteString(fmt.Sprintf("Query: %s\n", c.Request.URL.RawQuery))
	info.WriteString(fmt.Sprintf("Client IP: %s\n", c.ClientIP()))
	info.WriteString(fmt.Sprintf("User Agent: %s\n", c.Request.UserAgent()))
	info.WriteString(fmt.Sprintf("Content Type: %s\n", c.Request.Header.Get("Content-Type")))
	info.WriteString(fmt.Sprintf("Content Length: %d\n", c.Request.ContentLength))
	info.WriteString(fmt.Sprintf("Host: %s\n", c.Request.Host))
	info.WriteString(fmt.Sprintf("Remote Addr: %s\n", c.Request.RemoteAddr))
	info.WriteString(fmt.Sprintf("Request URI: %s\n", c.Request.RequestURI))
	info.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05")))

	// 添加重要的请求头
	importantHeaders := []string{
		"Authorization",
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Forwarded-Proto",
		"Accept",
		"Accept-Language",
		"Accept-Encoding",
		"Referer",
	}

	info.WriteString("Important Headers:\n")
	for _, header := range importantHeaders {
		if value := c.Request.Header.Get(header); value != "" {
			info.WriteString(fmt.Sprintf("  %s: %s\n", header, value))
		}
	}

	return info.String()
}

// logPanicError 记录panic错误日志
func logPanicError(logger Logger, ctx context.Context, errorMsg, requestInfo string, stack []byte, printStack bool) {
	// 构建完整的错误信息
	var fullErrorMsg strings.Builder

	fullErrorMsg.WriteString("=== PANIC RECOVERED ===\n")
	fullErrorMsg.WriteString(fmt.Sprintf("Error: %s\n", errorMsg))
	fullErrorMsg.WriteString("\n=== REQUEST INFORMATION ===\n")
	fullErrorMsg.WriteString(requestInfo)

	if len(stack) > 0 {
		fullErrorMsg.WriteString("\n=== STACK TRACE ===\n")
		fullErrorMsg.WriteString(string(stack))
	}

	fullErrorMsg.WriteString("\n=== END PANIC LOG ===\n")

	// 记录到日志
	logger.Error(ctx, "Panic Recovery", fullErrorMsg.String())

	// 如果启用了控制台打印，也打印到控制台
	if printStack {
		fmt.Print(fullErrorMsg.String())
	}
}

// RecoveryWithLogger 便捷方法，创建带日志的Recovery中间件
func RecoveryWithLogger(logger Logger) gin.HandlerFunc {
	return Recovery(&RecoveryConfig{
		Logger:           logger,
		EnableStackTrace: true,
		StackTraceSize:   4 << 10,
		PrintStack:       true,
	})
}

// RecoveryWithCustomHandler 便捷方法，创建带自定义处理器的Recovery中间件
func RecoveryWithCustomHandler(logger Logger, handler func(*gin.Context, interface{})) gin.HandlerFunc {
	return Recovery(&RecoveryConfig{
		Logger:           logger,
		EnableStackTrace: true,
		StackTraceSize:   4 << 10,
		PrintStack:       true,
		CustomHandler:    handler,
	})
}
