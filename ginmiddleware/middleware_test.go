package ginmiddleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ffhuo/go-kits/common/field"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mockLogger 模拟日志实现
type mockLogger struct {
	logs []string
}

func (m *mockLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	logEntry := m.buildLogEntry("INFO", msg, ctx, data...)
	m.logs = append(m.logs, logEntry)
}

func (m *mockLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logEntry := m.buildLogEntry("WARN", msg, ctx, data...)
	m.logs = append(m.logs, logEntry)
}

func (m *mockLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	logEntry := m.buildLogEntry("ERROR", msg, ctx, data...)
	m.logs = append(m.logs, logEntry)
}

func (m *mockLogger) Debug(ctx context.Context, msg string, data ...interface{}) {
	logEntry := m.buildLogEntry("DEBUG", msg, ctx, data...)
	m.logs = append(m.logs, logEntry)
}

func (m *mockLogger) buildLogEntry(level, msg string, ctx context.Context, data ...interface{}) string {
	logEntry := level + ": " + msg

	// 添加data参数
	if len(data) > 0 {
		logEntry += " " + fmt.Sprintf("%v", data[0])
	}

	// 获取field信息
	fields := field.Get(ctx)
	if len(fields) > 0 {
		var parts []string
		for _, f := range fields {
			key := f.Key()
			value := f.Value()

			// 特殊处理一些字段以匹配测试期望
			switch key {
			case "query":
				parts = append(parts, fmt.Sprintf("Query: %v", value))
			case "body":
				parts = append(parts, fmt.Sprintf("Body: %v", value))
			default:
				parts = append(parts, fmt.Sprintf("%s: %v", key, value))
			}
		}
		if len(parts) > 0 {
			logEntry += " | " + strings.Join(parts, " | ")
		}
	}

	return logEntry
}

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestID())

	r.GET("/test", func(c *gin.Context) {
		requestID := GetRequestID(c)
		if requestID == "" {
			t.Error("RequestID should not be empty")
		}
		c.JSON(200, gin.H{"request_id": requestID})
	})

	// 测试自动生成RequestID
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	requestID := w.Header().Get(RequestIDHeader)
	if requestID == "" {
		t.Error("Response should contain X-Request-ID header")
	}

	// 验证是否为有效的UUID
	if _, err := uuid.Parse(requestID); err != nil {
		t.Errorf("RequestID should be valid UUID, got: %s", requestID)
	}
}

func TestRequestIDFromHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestID())

	r.GET("/test", func(c *gin.Context) {
		requestID := GetRequestID(c)
		c.JSON(200, gin.H{"request_id": requestID})
	})

	// 测试从请求头获取RequestID
	expectedID := "test-request-id-123"
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, expectedID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	responseID := w.Header().Get(RequestIDHeader)
	if responseID != expectedID {
		t.Errorf("Expected RequestID %s, got %s", expectedID, responseID)
	}
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORS())

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// 测试普通请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS headers not set correctly")
	}
}

func TestCORSPreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORS())

	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// 测试预检请求
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("Expected status 204 for OPTIONS request, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Access-Control-Allow-Methods header should be set")
	}
}

func TestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := &mockLogger{}
	r := gin.New()
	r.Use(LoggerMiddleware(&LoggerConfig{
		Logger: logger,
	}))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if len(logger.logs) == 0 {
		t.Error("Logger should have recorded the request")
	}

	// 检查日志内容
	found := false
	for _, log := range logger.logs {
		if strings.Contains(log, "HTTP Request") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Logger should have recorded HTTP Request")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := &mockLogger{}
	r := gin.New()
	r.Use(RecoveryWithLogger(logger))

	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500 after panic, got %d", w.Code)
	}

	// 检查是否记录了panic日志
	found := false
	for _, log := range logger.logs {
		if strings.Contains(log, "Panic Recovery") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Logger should have recorded panic recovery")
	}
}

func TestRecoveryWithCustomHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := &mockLogger{}
	customHandlerCalled := false

	r := gin.New()
	r.Use(RecoveryWithCustomHandler(logger, func(c *gin.Context, err interface{}) {
		customHandlerCalled = true
		c.JSON(500, gin.H{"custom_error": "handled"})
	}))

	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500 after panic, got %d", w.Code)
	}

	if !customHandlerCalled {
		t.Error("Custom handler should have been called")
	}

	if !strings.Contains(w.Body.String(), "custom_error") {
		t.Error("Custom error response should be returned")
	}
}

func TestSkipPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := &mockLogger{}
	r := gin.New()
	r.Use(LoggerMiddleware(&LoggerConfig{
		Logger:    logger,
		SkipPaths: []string{"/health"},
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 检查是否跳过了日志记录
	if len(logger.logs) > 0 {
		t.Error("Logger should not have recorded requests to skip paths")
	}
}

func TestLoggerWithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := &mockLogger{}
	r := gin.New()
	r.Use(LoggerMiddleware(&LoggerConfig{
		Logger:          logger,
		ShowQueryParams: true,
	}))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test?name=john&age=25", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 检查日志是否包含query参数
	found := false
	for _, log := range logger.logs {
		if strings.Contains(log, "Query: name=john&age=25") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Logger should have recorded query parameters")
	}
}

func TestLoggerWithSensitiveQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := &mockLogger{}
	r := gin.New()
	r.Use(LoggerMiddleware(&LoggerConfig{
		Logger:          logger,
		ShowQueryParams: true,
		SensitiveFields: []string{"password", "token"},
	}))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test?username=john&password=secret123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 检查日志是否脱敏了敏感参数
	found := false
	for _, log := range logger.logs {
		if strings.Contains(log, "username=john") && strings.Contains(log, "password=***") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Logger should have masked sensitive query parameters")
	}
}

func TestLoggerWithRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := &mockLogger{}
	r := gin.New()
	r.Use(LoggerMiddleware(&LoggerConfig{
		Logger:          logger,
		ShowRequestBody: true,
		MaxBodySize:     1024,
	}))

	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	body := `{"name":"john","email":"john@example.com"}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 检查日志是否包含请求体
	found := false
	for _, log := range logger.logs {
		if strings.Contains(log, "Body: "+body) {
			found = true
			break
		}
	}
	if !found {
		t.Error("Logger should have recorded request body")
	}
}

func TestLoggerWithSensitiveRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := &mockLogger{}
	r := gin.New()
	r.Use(LoggerMiddleware(&LoggerConfig{
		Logger:          logger,
		ShowRequestBody: true,
		MaxBodySize:     1024,
		SensitiveFields: []string{"password", "secret"},
	}))

	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	body := `{"username":"john","password":"secret123"}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 检查日志是否脱敏了敏感字段
	found := false
	for _, log := range logger.logs {
		if strings.Contains(log, `"username":"john"`) && strings.Contains(log, `"password":"***"`) {
			found = true
			break
		}
	}
	if !found {
		t.Error("Logger should have masked sensitive fields in request body")
	}
}

func TestLoggerSkipBodyMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := &mockLogger{}
	r := gin.New()
	r.Use(LoggerMiddleware(&LoggerConfig{
		Logger:          logger,
		ShowRequestBody: true,
		SkipBodyMethods: []string{"GET", "HEAD"},
	}))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", strings.NewReader(`{"data":"test"}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 检查日志不应该包含请求体
	for _, log := range logger.logs {
		if strings.Contains(log, "Body:") {
			t.Error("Logger should not record body for GET requests when configured to skip")
		}
	}
}

func TestLoggerLargeBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := &mockLogger{}
	r := gin.New()
	r.Use(LoggerMiddleware(&LoggerConfig{
		Logger:          logger,
		ShowRequestBody: true,
		MaxBodySize:     10, // 限制为10字节
	}))

	r.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	body := `{"name":"john","email":"john@example.com","description":"very long description"}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 检查日志是否截断了大的请求体
	found := false
	for _, log := range logger.logs {
		if strings.Contains(log, "Body: ") && strings.Contains(log, "...") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Logger should have truncated large request body")
	}
}

func TestSimpleLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := &mockLogger{}
	r := gin.New()
	r.Use(SimpleLoggerMiddleware(logger))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test?param=value", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 检查简单日志是否记录了基本信息
	found := false
	for _, log := range logger.logs {
		if strings.Contains(log, "HTTP Request") &&
			strings.Contains(log, "method: GET") &&
			strings.Contains(log, "path: /test") &&
			strings.Contains(log, "status: 200") {
			found = true
			// 简单日志不应该包含query参数
			if strings.Contains(log, "param=value") {
				t.Error("Simple logger should not include query parameters")
			}
			break
		}
	}
	if !found {
		t.Error("Simple logger should have recorded basic request information")
	}
}
