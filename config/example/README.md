# 配置包使用示例

这个示例展示了如何使用 go-kits/config 包来管理应用程序配置。

## 功能特点

1. 支持多种配置文件格式（YAML、TOML、JSON等）
2. 支持配置文件合并
3. 支持环境变量覆盖
4. 支持配置文件热重载
5. 支持结构化配置

## 目录结构

```
.
├── README.md
├── configs/
│   ├── config.yaml          # 主配置文件
│   └── config.local.yaml    # 本地开发配置文件
└── main.go                  # 示例代码
```

## 使用方法

1. 创建配置实例：
```go
cfg := config.New().
    AddPath("configs").      // 添加配置文件路径
    SetEnvPrefix("APP").     // 设置环境变量前缀
    AutomaticEnv()          // 启用自动环境变量绑定
```

2. 加载配置文件：
```go
err := cfg.LoadFile("config.yaml")
```

3. 合并其他配置文件：
```go
err := cfg.MergeFile("config.local.yaml")
```

4. 获取配置值：
```go
var mysqlConfig config.MySQLConfig
err := cfg.Unmarshal("mysql", &mysqlConfig)
```

5. 监听配置变更：
```go
cfg.Watch(func(event fsnotify.Event) {
    if event.Op&fsnotify.Write == fsnotify.Write {
        // 处理配置更新
    }
})
```

## 环境变量

可以使用环境变量覆盖配置文件中的值。环境变量名称格式为：`APP_SECTION_KEY`

例如：
- `APP_MYSQL_DSN`：覆盖 MySQL DSN
- `APP_REDIS_PASSWORD`：覆盖 Redis 密码
- `APP_MONGODB_URI`：覆盖 MongoDB URI

## 运行示例

```bash
go run main.go
```

修改 `configs/config.yaml` 或 `configs/config.local.yaml` 文件，观察配置热重载效果。
