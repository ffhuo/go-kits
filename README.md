# Go-Kits

Go-Kits 是一个用于 Go 项目的工具包集合，提供了多种常用功能组件，帮助开发者快速构建高质量的 Go 应用程序。

## 项目结构

Go-Kits 由多个独立的包组成，每个包都专注于特定的功能领域：

### 核心包

- **core**: 提供 Web 服务核心功能，基于 Gin 框架封装，支持中间件、Swagger、Prometheus 监控等特性。
- **common**: 通用工具集合，包含多个子包：
  - `bar`: 进度条实现
  - `conversion`: 数据类型转换工具
  - `crypto`: 加密解密工具
  - `ctime`: 时间处理工具
  - `errno`: 错误码定义
  - `paginator`: 分页工具
  - `shutdown`: 优雅关闭服务
  - `utils`: 通用工具函数
  - `validation`: 数据验证工具
  - `x509cert`: X509 证书工具

### 数据存储与缓存

- **sqldb**: 数据库操作工具，提供 SQL 执行追踪和指标收集功能。
- **redis**: Redis 客户端封装，支持单机和集群模式，提供常用操作方法。
- **influxdb**: InfluxDB 时序数据库客户端封装。

### 配置与日志

- **config**: 配置管理工具，基于 Viper 封装，支持多种配置源和动态配置更新。
- **logger**: 高性能结构化日志库，基于 Zap 封装，支持字段日志、SQL 追踪和文件轮转等特性。

### 消息与通信

- **message**: 消息处理工具。
- **mqtt**: MQTT 客户端封装。
- **gout**: HTTP 客户端工具。

### 监控与服务发现

- **prometheus**: Prometheus 监控工具，提供指标收集和导出功能。
- **etcd**: ETCD 客户端封装，用于服务发现和配置管理。

### 数据处理

- **excel**: Excel 文件处理工具，基于 excelize 封装，提供简化的 Excel 读写功能。

## 安装与使用

Go-Kits 采用 Go Modules 进行依赖管理，可以按需引入所需的包：

```bash
# 安装特定包
go get github.com/ffhuo/go-kits/logger
go get github.com/ffhuo/go-kits/config
go get github.com/ffhuo/go-kits/redis
# 等等...
```

## 包使用示例

### Logger 示例

```go
package main

import (
    "context"
    "github.com/ffhuo/go-kits/logger"
)

func main() {
    log, err := logger.New(
        logger.WithDebugLevel(),
        logger.WithFormatJSON(),
        logger.WithFileRotationP(
            "logs/app.log",
            100,
            7,
            10,
        ),
        logger.WithField("app", "example"),
    )
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    log.Info(ctx, "Hello, %s!", "World")
}
```

### Config 示例

```go
package main

import (
    "fmt"
    "github.com/ffhuo/go-kits/config"
    "github.com/fsnotify/fsnotify"
)

func main() {
    cfg := config.New()
    cfg.AddPath(".")
    cfg.SetEnvPrefix("APP")
    cfg.AutomaticEnv()
    
    err := cfg.LoadFile("config.yaml")
    if err != nil {
        panic(err)
    }
    
    // 获取配置值
    port := cfg.GetInt("server.port")
    fmt.Println("Server port:", port)
    
    // 监听配置变化
    cfg.Watch(func(e fsnotify.Event) {
        fmt.Println("Config file changed:", e.Name)
    })
}
```

### Redis 示例

```go
package main

import (
    "context"
    "fmt"
    "github.com/ffhuo/go-kits/redis"
    "time"
)

func main() {
    ctx := context.Background()
    client, err := redis.New(ctx, []string{"localhost:6379"}, "")
    if err != nil {
        panic(err)
    }
    
    // 设置值
    err = client.Set("key", "value", time.Hour)
    if err != nil {
        panic(err)
    }
    
    // 获取值
    value, err := client.Get("key")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Value:", value)
}
```

### Excel 示例

```go
package main

import (
    "fmt"
    "github.com/ffhuo/go-kits/excel"
)

func main() {
    // 创建新的 Excel 文件
    xl := excel.New("Sheet1")
    
    // 写入表格数据
    headers := []string{"Name", "Age", "Email"}
    data := [][]interface{}{
        {"张三", 30, "zhangsan@example.com"},
        {"李四", 25, "lisi@example.com"},
    }
    
    err := xl.WriteTable(headers, data)
    if err != nil {
        panic(err)
    }
    
    // 保存文件
    err = xl.SaveAs("users.xlsx")
    if err != nil {
        panic(err)
    }
}
```

## 待优化

- **excel 包**: 当前实现较为复杂，计划进行重构以提供更简洁、易用的 API。

## 贡献

欢迎提交 Issues 和 Pull Requests。

## 许可证

MIT 许可证