# Logger V2

基于 Go 标准库 `log/slog` 的日志组件，提供结构化日志记录功能。

## 特性

- 基于 Go 1.21+ 标准库 `log/slog`
- 支持多种日志级别（Debug、Info、Warn、Error）
- 支持 JSON 和文本格式输出
- 支持文件输出和日志轮转
- 支持上下文字段
- 兼容 GORM 日志接口
- 支持源码位置信息
- 支持自定义时间格式
- 支持日志分组

## 安装

```bash
go get github.com/go-kits/logger_v2
```

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "github.com/go-kits/logger_v2"
)

func main() {
    // 创建默认日志器
    logger, err := logger_v2.New()
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    logger.Info(ctx, "Hello, World!")
}
```

### 自定义配置

```go
// 使用选项模式
logger, err := logger_v2.New(
    logger_v2.WithDebugLevel(),
    logger_v2.WithFormatJSON(),
    logger_v2.WithSource(),
    logger_v2.WithField("service", "my-app"),
    logger_v2.WithFileRotationP("logs/app.log", 10, 5, 7),
)

// 使用配置结构
config := logger_v2.Config{
    Level:      logger_v2.InfoLevel,
    Format:     "json",
    FilePath:   "logs/app.log",
    AddSource:  true,
    Fields: map[string]string{
        "app": "my-app",
        "env": "production",
    },
}
logger, err := logger_v2.NewFromConfig(config)
```

## API 参考

### 创建日志器

#### 选项函数

- `WithDebugLevel()` - 设置 Debug 级别
- `WithInfoLevel()` - 设置 Info 级别
- `WithWarnLevel()` - 设置 Warn 级别
- `WithErrorLevel()` - 设置 Error 级别
- `WithFormatJSON()` - 使用 JSON 格式
- `WithSource()` - 添加源码位置信息
- `WithField(key, value)` - 添加固定字段
- `WithFileP(file)` - 输出到文件
- `WithFileRotationP(file, maxSize, maxBackups, maxAge)` - 文件轮转
- `WithTimeLayout(layout)` - 自定义时间格式
- `WithDisableConsole()` - 禁用控制台输出

#### 配置结构

```go
type Config struct {
    Level          LogLevel          `json:"level"`
    Format         string            `json:"format"`
    FilePath       string            `json:"file_path"`
    MaxSize        int              `json:"max_size"`
    MaxAge         int              `json:"max_age"`
    MaxBackups     int              `json:"max_backups"`
    DisableConsole bool             `json:"disable_console"`
    TimeLayout     string           `json:"time_layout"`
    AddSource      bool             `json:"add_source"`
    Fields         map[string]string `json:"fields"`
}
```

### 日志记录

```go
// 基本日志记录
logger.Debug(ctx, "Debug message")
logger.Info(ctx, "Info message")
logger.Warn(ctx, "Warning message")
logger.Error(ctx, "Error message")

// 格式化消息
logger.Info(ctx, "User %s logged in", username)

// 结构化日志
logger.Info(ctx, "User action",
    slog.String("action", "login"),
    slog.String("username", "john"),
    slog.Int("attempt", 1),
)
```

### 上下文字段

```go
// 添加上下文字段
ctx = logger_v2.WithFields(ctx,
    logger_v2.NewMeta("request_id", "12345"),
    logger_v2.NewMeta("user_id", "user123"),
)

// 使用带字段的上下文
logger.Info(ctx, "Processing request") // 会自动包含上下文字段
```

### 日志器扩展

```go
// 添加属性
dbLogger := logger.With(
    slog.String("component", "database"),
    slog.Int("pool_size", 10),
)

// 添加分组
httpLogger := logger.WithGroup("http")
httpLogger.Info(ctx, "Request processed",
    slog.String("method", "GET"),
    slog.Int("status", 200),
)
```

### 级别检查

```go
if logger.DebugEnabled() {
    logger.Debug(ctx, "Expensive debug operation")
}
```

## GORM 集成

logger_v2 实现了 GORM 的日志接口，可以直接用于 GORM：

```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "github.com/go-kits/logger_v2"
)

logger, _ := logger_v2.New(logger_v2.WithDebugLevel())

db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger,
})
```

## 性能考虑

- 基于 `log/slog`，性能优于传统的格式化日志
- 支持零分配的结构化日志记录
- 内置级别检查避免不必要的计算
- 文件轮转使用 lumberjack，性能稳定

## 与 logger v1 的区别

| 特性 | logger v1 | logger_v2 |
|------|-----------|-----------|
| 底层实现 | zap | slog (标准库) |
| Go 版本要求 | 1.16+ | 1.21+ |
| 性能 | 极高 | 高 |
| 标准化 | 第三方 | 官方标准 |
| 属性类型 | zap.Field | slog.Attr |
| 上下文支持 | 有限 | 原生支持 |

## 迁移指南

从 logger v1 迁移到 logger_v2：

1. 更新导入路径
2. 替换 `zap.Field` 为 `slog.Attr`
3. 使用 `slog.String()`, `slog.Int()` 等函数
4. 更新上下文字段处理

## 示例

查看 `example/main.go` 获取更多使用示例。

## 许可证

MIT License 