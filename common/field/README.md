# Field Package

简洁的日志字段管理包，用于在context中传递结构化日志字段。

## 特性

- 简洁的API：`F()` 创建字段，`With()` 添加字段，`Get()` 获取字段
- 支持gin.Context和标准context.Context
- 与logger和logger_v2包无缝集成
- 零依赖，纯Go实现

## 使用方法

### 基本用法

```go
import "github.com/ffhuo/go-kits/common/field"

// 创建字段
userID := field.F("user_id", "12345")
requestID := field.F("request_id", "req-abc")

// 添加到context
ctx = field.With(ctx, userID, requestID)

// 获取字段
fields := field.Get(ctx)
```

### 在中间件中使用

```go
func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 添加请求相关字段
        c = field.With(c, 
            field.F("ip", c.ClientIP()),
            field.F("method", c.Request.Method),
            field.F("path", c.Request.URL.Path),
            field.F("agent", c.Request.UserAgent()),
        )
        
        c.Next()
    }
}
```

### 在业务逻辑中使用

```go
func UserService(ctx context.Context) {
    // 添加业务字段
    ctx = field.With(ctx, 
        field.F("service", "user"),
        field.F("action", "login"),
    )
    
    // 使用logger打印（自动包含所有字段）
    logger.Info(ctx, "user login successful")
}
```

## 与Logger集成

字段会自动被logger包识别和格式化：

- `logger` 包：转换为zap.Field格式
- `logger_v2` 包：转换为slog格式

无需手动转换，直接使用即可。 