# Logger

基于 [zap](https://github.com/uber-go/zap) 的高性能结构化日志库，支持字段日志、SQL追踪和文件轮转等特性。

## 特性

- 多级日志支持（Debug、Info、Warn、Error）
- 结构化日志，支持字段注入
- SQL 查询追踪，包含执行时间
- 文件轮转
- 支持 JSON 和控制台输出格式
- 自定义时间格式
- 基于 Context 的字段注入
- 高性能（基于 zap）

## 安装

```bash
go get github.com/ffhuo/go-kits/logger
```

## 快速开始

```go
package main

import (
    "context"
    "time"
    "github.com/ffhuo/go-kits/logger"
)

func main() {
    log, err := logger.New(
        logger.WithDebugLevel(),         // 设置日志级别
        logger.WithFormatJSON(),         // 使用 JSON 格式
        logger.WithFileRotationP(        // 启用文件轮转
            "logs/app.log",              // 日志文件路径
            100,                         // 单个文件最大尺寸（MB）
            7,                           // 保留天数
            10,                          // 最大备份数量
        ),
        logger.WithField("app", "example"), // 添加全局字段
    )
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    log.Info(ctx, "你好，%s！", "世界")
}
```

## 配置选项

- `WithDebugLevel()`: 设置日志级别为 debug
- `WithInfoLevel()`: 设置日志级别为 info
- `WithWarnLevel()`: 设置日志级别为 warn
- `WithErrorLevel()`: 设置日志级别为 error
- `WithFormatJSON()`: 使用 JSON 输出格式
- `WithTimeLayout(layout string)`: 设置时间格式
- `WithField(key, value string)`: 添加全局字段
- `WithFileP(filename string)`: 输出到文件
- `WithFileRotationP(filename string, maxSize, maxAge, maxBackups int)`: 启用文件轮转
- `WithDisableConsole()`: 禁用控制台输出

## 字段日志

```go
// 添加请求相关字段
ctx = logger.WithFields(ctx,
    logger.NewMeta("requestID", "123456"),
    logger.NewMeta("userID", "user123"),
)
log.Info(ctx, "收到请求")

// 添加结构化数据
ctx = logger.WithFields(ctx,
    logger.NewMeta("user", map[string]interface{}{
        "id": "user123",
        "name": "张三",
        "role": "admin",
    }),
)
log.Info(ctx, "用户信息")
```

## SQL 追踪

```go
// 追踪 SQL 执行
begin := time.Now()
log.Trace(ctx, begin, func() (string, int64) {
    return "SELECT * FROM users WHERE id = ?", 1
}, nil)

// 带错误的 SQL 追踪
log.Trace(ctx, begin, func() (string, int64) {
    return "INSERT INTO users (name) VALUES (?)", 0
}, fmt.Errorf("插入用户失败：%s", "张三"))
```

## 输出示例

```json
{
    "level": "info",
    "time": "2024-12-27T11:23:48+08:00",
    "caller": "example/main.go:31",
    "msg": "收到请求",
    "app": "example",
    "requestID": "123456",
    "userID": "user123"
}
```

## 在 Gin 中使用

```go
func main() {
    r := gin.New()
    
    // 使用日志中间件
    r.Use(func(c *gin.Context) {
        ctx := logger.WithFields(c,
            logger.NewMeta("requestID", c.GetHeader("X-Request-ID")),
            logger.NewMeta("clientIP", c.ClientIP()),
        )
        c.Next()
    })
    
    r.GET("/ping", func(c *gin.Context) {
        log.Info(c, "收到 ping 请求")
        c.JSON(200, gin.H{"message": "pong"})
    })
}
```

## 贡献

欢迎提交 issues 和 pull requests。

## 许可证

MIT 许可证
