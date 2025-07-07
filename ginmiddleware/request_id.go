package ginmiddleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDKey 请求ID在上下文中的键名
	RequestIDKey = "request_id"
	// RequestIDHeader 请求ID在HTTP头中的键名
	RequestIDHeader = "X-Request-ID"
)

// RequestID 中间件，为每个请求生成唯一的RequestID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从请求头中获取RequestID
		requestID := c.GetHeader(RequestIDHeader)

		// 如果请求头中没有RequestID，则生成一个新的
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置到gin上下文中
		c.Set(RequestIDKey, requestID)

		// 设置到响应头中
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID 从gin上下文中获取RequestID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		return requestID.(string)
	}
	return ""
}
