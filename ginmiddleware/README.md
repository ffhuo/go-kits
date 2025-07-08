# Gin Middleware

一个功能完整的Gin中间件包，提供常用的HTTP中间件功能。

## 功能特性

- **RequestID**: 为每个请求生成唯一的请求ID
- **CORS**: 跨域资源共享支持
- **Logger**: 详细的请求日志记录
- **Recovery**: Panic恢复和错误处理

## 安装

```bash
go get github.com/ffhuo/go-kits/ginmiddleware
```

## 使用方法

### 基本使用

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/ffhuo/go-kits/ginmiddleware"
    "github.com/ffhuo/go-kits/logger" // 或 logger_v2
)

func main() {
    r := gin.New()
    
    // 创建日志实例
    log, _ := logger.New()
    
    // 添加中间件
    r.Use(ginmiddleware.RequestID())
    r.Use(ginmiddleware.CORS())
    r.Use(ginmiddleware.LoggerMiddleware(&ginmiddleware.LoggerConfig{
        Logger: log,
    }))
    r.Use(ginmiddleware.RecoveryWithLogger(log))
    
    r.GET("/", func(c *gin.Context) {
        requestID := ginmiddleware.GetRequestID(c)
        c.JSON(200, gin.H{
            "message": "Hello World",
            "request_id": requestID,
        })
    })
    
    r.Run(":8080")
}
```

## 中间件详解

### 1. RequestID 中间件

为每个请求生成唯一的请求ID，支持从请求头中获取已有的ID。

```go
// 基本使用
r.Use(ginmiddleware.RequestID())

// 获取请求ID
func handler(c *gin.Context) {
    requestID := ginmiddleware.GetRequestID(c)
    // 使用 requestID...
}
```

**特性：**
- 自动生成UUID作为请求ID
- 优先使用请求头中的 `X-Request-ID`
- 自动添加到响应头中
- 提供便捷的获取方法

### 2. CORS 中间件

处理跨域请求，支持预检请求和灵活的配置。

```go
// 使用默认配置
r.Use(ginmiddleware.CORS())

// 自定义配置
r.Use(ginmiddleware.CORS(&ginmiddleware.CORSConfig{
    AllowOrigins:     []string{"https://example.com", "https://app.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           86400,
}))
```

**特性：**
- 支持多个允许的源
- 灵活的方法和头部配置
- 预检请求处理
- 凭证支持

### 3. Logger 中间件

详细的请求日志记录，兼容 logger 和 logger_v2，支持可控制的参数显示和敏感数据脱敏。

```go
// 基本使用
r.Use(ginmiddleware.LoggerMiddleware(&ginmiddleware.LoggerConfig{
    Logger: yourLogger,
}))

// 自定义配置
r.Use(ginmiddleware.LoggerMiddleware(&ginmiddleware.LoggerConfig{
    Logger:           yourLogger,
    SkipPaths:        []string{"/health", "/metrics"},
    SlowThreshold:    500 * time.Millisecond,
    ShowQueryParams:  true,  // 显示GET请求的query参数
    ShowRequestBody:  true,  // 显示POST/PUT/PATCH请求的body内容
    MaxBodySize:      1024,  // 最大body显示大小（字节）
    SkipBodyMethods:  []string{"GET", "HEAD", "OPTIONS"}, // 跳过显示body的方法
    SensitiveFields:  []string{"password", "token", "secret"}, // 敏感字段脱敏
}))

// 简化版日志中间件（只记录基本信息）
r.Use(ginmiddleware.SimpleLoggerMiddleware(yourLogger))
```

**记录信息：**
- 请求ID
- HTTP方法和路径
- 状态码和响应时间
- 客户端IP和User-Agent（截取前100字符）
- GET请求的query参数（可控制）
- POST/PUT/PATCH请求的body内容（可控制）
- 敏感数据自动脱敏
- 错误信息

**特性：**
- **智能参数显示**: GET请求显示query参数，POST请求显示body内容
- **敏感数据脱敏**: 自动识别并脱敏敏感字段（如password、token等）
- **大小限制**: 支持限制body显示大小，避免日志过大
- **方法过滤**: 可配置跳过特定HTTP方法的body记录
- **二进制检测**: 自动检测并标识二进制内容
- **性能优化**: 移除了重复的时间记录（logger本身已有时间）

### 4. Recovery 中间件

处理panic并记录详细的错误信息和堆栈跟踪。

```go
// 基本使用
r.Use(ginmiddleware.RecoveryWithLogger(yourLogger))

// 自定义配置
r.Use(ginmiddleware.Recovery(&ginmiddleware.RecoveryConfig{
    Logger:           yourLogger,
    EnableStackTrace: true,
    PrintStack:       true,
    CustomHandler: func(c *gin.Context, err interface{}) {
        // 自定义错误处理
        c.JSON(500, gin.H{"error": "Something went wrong"})
    },
}))
```

**错误信息包含：**
- 详细的请求信息
- 完整的堆栈跟踪
- 错误发生时间
- 请求ID关联
- 格式化的错误输出

## 日志接口

中间件使用通用的日志接口，兼容多种日志实现：

```go
type Logger interface {
    Info(ctx context.Context, msg string, data ...interface{})
    Warn(ctx context.Context, msg string, data ...interface{})
    Error(ctx context.Context, msg string, data ...interface{})
    Debug(ctx context.Context, msg string, data ...interface{})
}
```

支持的日志实现：
- `github.com/ffhuo/go-kits/logger` (基于zap)
- `github.com/ffhuo/go-kits/logger_v2` (基于slog)

## 完整示例

```go
package main

import (
    "context"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/ffhuo/go-kits/ginmiddleware"
    "github.com/ffhuo/go-kits/logger"
)

func main() {
    // 创建日志实例
    log, err := logger.New(
        logger.WithInfoLevel(),
        logger.WithFormatJSON(),
        logger.WithFileRotationP("logs/app.log", 100, 7, 30),
    )
    if err != nil {
        panic(err)
    }

    // 创建Gin实例
    r := gin.New()

    // 添加中间件
    r.Use(ginmiddleware.RequestID())
    r.Use(ginmiddleware.CORS(&ginmiddleware.CORSConfig{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"*"},
        AllowCredentials: false,
    }))
    r.Use(ginmiddleware.LoggerMiddleware(&ginmiddleware.LoggerConfig{
        Logger:        log,
        SkipPaths:     []string{"/health"},
        SlowThreshold: 200 * time.Millisecond,
    }))
    r.Use(ginmiddleware.RecoveryWithLogger(log))

    // 路由
    r.GET("/", func(c *gin.Context) {
        requestID := ginmiddleware.GetRequestID(c)
        c.JSON(200, gin.H{
            "message":    "Hello World",
            "request_id": requestID,
            "timestamp":  time.Now(),
        })
    })

    r.GET("/panic", func(c *gin.Context) {
        panic("这是一个测试panic")
    })

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // 启动服务器
    r.Run(":8080")
}
```

## 最佳实践

1. **中间件顺序**: 建议按照 RequestID -> CORS -> Logger -> Recovery 的顺序添加中间件
2. **日志配置**: 根据环境选择合适的日志级别和输出格式
3. **错误处理**: 使用自定义错误处理器来提供更好的用户体验
4. **性能考虑**: 合理配置跳过路径和慢请求阈值

## 许可证

MIT License 