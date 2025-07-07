package ginmiddleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// mockLogger 模拟日志实现
type mockLogger struct {
	logs []string
}

func (m *mockLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	m.logs = append(m.logs, "INFO: "+msg)
}

func (m *mockLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	m.logs = append(m.logs, "WARN: "+msg)
}

func (m *mockLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	m.logs = append(m.logs, "ERROR: "+msg)
}

func (m *mockLogger) Debug(ctx context.Context, msg string, data ...interface{}) {
	m.logs = append(m.logs, "DEBUG: "+msg)
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
