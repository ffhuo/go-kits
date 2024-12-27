# SQL Database Module

这是一个基于 GORM 的数据库操作模块，提供了统一的数据库操作接口，支持多种数据库（MySQL、SQLite、ClickHouse），并集成了监控、追踪和日志功能。

## 特性

- 支持多种数据库
  - MySQL
  - SQLite
  - ClickHouse
- 读写分离支持
- 完整的监控指标
  - SQL 执行时间
  - 影响行数统计
  - 错误计数
  - 慢查询监控
- 结构化日志
- Prometheus 集成
- 分页查询支持

## 安装

```bash
go get github.com/ffhuo/go-kits/sqldb
```

## 快速开始

### 基本用法

```go
package main

import (
    "context"
    "time"
    
    "github.com/ffhuo/go-kits/logger"
    "github.com/ffhuo/go-kits/sqldb"
    "github.com/ffhuo/go-kits/paginator"
)

// User 用户模型
type User struct {
    ID        uint      `gorm:"primarykey"`
    Name      string
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
}

func main() {
    // 创建日志实例
    log := logger.NewLogger()
    
    // 创建数据库配置
    opt := sqldb.NewOptions().
        WithMySQL("user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local").
        WithLogger(log)
    
    // 创建数据库实例
    db, err := sqldb.New(opt)
    if err != nil {
        panic(err)
    }
    
    // 创建记录
    user := &User{
        Name:  "John Doe",
        Email: "john@example.com",
    }
    if err := db.Create(context.Background(), user); err != nil {
        panic(err)
    }
    
    // 查询单条记录
    result := new(User)
    if err := db.First(context.Background(), result, sqldb.Where("name = ?", "John Doe")); err != nil {
        panic(err)
    }
    
    // 分页查询
    users := make([]*User, 0)
    page := paginator.NewPaginator(1, 10)
    total, err := db.Find(context.Background(), &users, page, sqldb.Where("name LIKE ?", "John%"))
    if err != nil {
        panic(err)
    }
    
    // 更新记录
    if err := db.Update(context.Background(), user, sqldb.Set("name", "Jane Doe")); err != nil {
        panic(err)
    }
    
    // 删除记录
    if err := db.Delete(context.Background(), user); err != nil {
        panic(err)
    }
}
```

### 读写分离

```go
opt := sqldb.NewOptions().
    WithMySQL("user:password@tcp(master:3306)/dbname").
    WithSlaves([]string{
        "user:password@tcp(slave1:3306)/dbname",
        "user:password@tcp(slave2:3306)/dbname",
    })
```

### 监控集成

```go
// 创建指标收集器（设置慢查询阈值为 200ms）
metrics := sqldb.NewSQLMetrics(200 * time.Millisecond)

// 创建追踪插件
trace := sqldb.NewTrace(metrics, log)

// 注册到数据库
db.Use(trace)
```

### 查询构建器

```go
// 条件查询
users := make([]*User, 0)
db.Find(context.Background(), &users,
    sqldb.Where("age > ?", 18),
    sqldb.Order("created_at DESC"),
    sqldb.Limit(10),
)

// 复杂查询
db.Find(context.Background(), &users,
    sqldb.Select("name, email"),
    sqldb.Where("age BETWEEN ? AND ?", 18, 30),
    sqldb.Or("vip = ?", true),
    sqldb.Group("department"),
    sqldb.Having("count(*) > ?", 5),
)
```

## Prometheus 指标

模块提供以下 Prometheus 指标：

- `sql_duration_seconds`: SQL 执行时间直方图
- `sql_errors_total`: SQL 错误计数器
- `sql_rows`: 影响行数直方图
- `sql_slow_queries_total`: 慢查询计数器

## 配置选项

```go
opt := sqldb.NewOptions().
    WithMySQL(dsn).                    // 设置 MySQL 连接
    WithSQLite(path).                  // 设置 SQLite 连接
    WithClickHouse(dsn).               // 设置 ClickHouse 连接
    WithLogger(log).                   // 设置日志
    WithSlowThreshold(200 * time.Millisecond).  // 设置慢查询阈值
    WithMaxIdleConns(10).              // 设置最大空闲连接数
    WithMaxOpenConns(100).             // 设置最大打开连接数
    WithConnMaxLifetime(time.Hour)     // 设置连接最大生命周期
```

## 注意事项

1. 在生产环境中，建议：
   - 设置适当的连接池参数
   - 配置慢查询阈值
   - 启用 Prometheus 监控
   - 使用结构化日志

2. 读写分离时：
   - 确保主从数据库配置正确
   - 注意数据一致性问题
   - 合理使用事务

3. 性能优化：
   - 合理使用索引
   - 避免大事务
   - 控制查询结果集大小
   - 使用适当的批处理大小

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个模块。