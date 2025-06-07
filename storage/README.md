# Storage - 文件存储管理包

Storage 是一个统一的文件存储管理包，支持多种存储后端，包括本地存储、数据库存储和腾讯云COS存储。

## 特性

- 🗂️ **统一接口**: 提供统一的存储接口，支持多种存储后端
- 💾 **本地存储**: 支持本地文件系统存储
- 🗄️ **数据库存储**: 支持将文件存储在数据库中（MySQL、PostgreSQL、SQLite）
- ☁️ **云存储**: 支持腾讯云COS对象存储
- 🔧 **存储管理器**: 支持多存储实例管理
- 🏭 **工厂模式**: 通过配置文件创建存储实例
- 📊 **文件元数据**: 支持文件元数据管理
- 🔗 **URL生成**: 支持生成文件访问URL
- 🗃️ **可选数据库元数据**: 所有存储方式都可选择使用数据库存储文件元数据信息

## 安装

```bash
go get github.com/ffhuo/go-kits/storage
```

## 快速开始

### 本地存储

```go
package main

import (
    "context"
    "strings"
    "github.com/ffhuo/go-kits/storage"
)

func main() {
    // 创建本地存储配置
    config := &storage.LocalConfig{
        RootPath: "/tmp/storage",
        BaseURL:  "http://localhost:8080/files",
    }

    // 创建存储实例
    store, err := storage.NewLocalStorage(config)
    if err != nil {
        panic(err)
    }
    defer store.Close()

    ctx := context.Background()

    // 上传文件
    content := "Hello, World!"
    reader := strings.NewReader(content)
    opts := &storage.UploadOptions{
        ContentType: "text/plain",
        Metadata: map[string]string{
            "author": "张三",
        },
    }

    fileInfo, err := store.Upload(ctx, "documents/hello.txt", reader, opts)
    if err != nil {
        panic(err)
    }

    // 下载文件
    downloadReader, err := store.Download(ctx, "documents/hello.txt")
    if err != nil {
        panic(err)
    }
    defer downloadReader.Close()

    // 检查文件是否存在
    exists, err := store.Exists(ctx, "documents/hello.txt")
    if err != nil {
        panic(err)
    }
}
```

### 数据库存储

```go
// 创建数据库存储配置
config := &storage.DBConfig{
    DSN:       "user:password@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True&loc=Local",
    TableName: "file_storage",
}

// 创建存储实例
store, err := storage.NewDatabaseStorage(config)
if err != nil {
    panic(err)
}
defer store.Close()
```

### 腾讯云COS存储

```go
// 创建COS存储配置
config := &storage.COSConfig{
    SecretID:  "your-secret-id",
    SecretKey: "your-secret-key",
    Region:    "ap-beijing",
    Bucket:    "your-bucket-name",
    BaseURL:   "https://your-custom-domain.com", // 可选
}

// 创建存储实例
store, err := storage.NewCOSStorage(config)
if err != nil {
    panic(err)
}
defer store.Close()
```

### 使用工厂函数

```go
// 通过配置创建存储实例
config := &storage.Config{
    Type: storage.StorageTypeLocal,
    Settings: map[string]interface{}{
        "rootPath": "/tmp/storage",
        "baseURL":  "http://localhost:8080/files",
    },
}

store, err := storage.NewStorage(config)
if err != nil {
    panic(err)
}
defer store.Close()
```

### 使用存储管理器

```go
// 创建存储管理器
manager := storage.NewStorageManager()
defer manager.Close()

// 注册多个存储实例
manager.Register("local", localStorage)
manager.Register("database", databaseStorage)
manager.Register("cos", cosStorage)

// 获取存储实例
store, err := manager.Get("local")
if err != nil {
    panic(err)
}

// 使用存储实例
fileInfo, err := store.Upload(ctx, "test.txt", reader, nil)
```

## 接口说明

### Storage 接口

```go
type Storage interface {
    // Upload 上传文件
    Upload(ctx context.Context, path string, reader io.Reader, opts *UploadOptions) (*FileInfo, error)
    
    // Download 下载文件
    Download(ctx context.Context, path string) (io.ReadCloser, error)
    
    // Delete 删除文件
    Delete(ctx context.Context, path string) error
    
    // Exists 检查文件是否存在
    Exists(ctx context.Context, path string) (bool, error)
    
    // GetInfo 获取文件信息
    GetInfo(ctx context.Context, path string) (*FileInfo, error)
    
    // List 列出文件
    List(ctx context.Context, opts *ListOptions) ([]*FileInfo, error)
    
    // GetURL 获取文件访问URL
    GetURL(ctx context.Context, path string, expiry time.Duration) (string, error)
    
    // Copy 复制文件
    Copy(ctx context.Context, srcPath, dstPath string) error
    
    // Move 移动文件
    Move(ctx context.Context, srcPath, dstPath string) error
    
    // Close 关闭存储连接
    Close() error
}
```

### 文件信息结构

```go
type FileInfo struct {
	ID          string            `gorm:"primaryKey;size:64" json:"id"`              // 文件唯一标识
	Name        string            `gorm:"size:255;not null" json:"name"`             // 文件名
	Path        string            `gorm:"size:500;not null;uniqueIndex" json:"path"` // 文件路径
	Size        int64             `gorm:"not null" json:"size"`                      // 文件大小（字节）
	ContentType string            `gorm:"size:100" json:"contentType"`               // 文件MIME类型
	Hash        string            `gorm:"size:64" json:"hash"`                       // 文件哈希值
	StorageType string            `gorm:"size:20" json:"storageType"`                // 存储类型（可选，用于数据库存储时区分）
	CreatedAt   time.Time         `json:"createdAt"`                                 // 创建时间
	UpdatedAt   time.Time         `json:"updatedAt"`                                 // 更新时间
	Metadata    map[string]string `gorm:"serializer:json" json:"metadata"`           // 元数据
}
```

**注意**: `FileInfo` 结构体现在包含了 GORM 标签，可以直接用于数据库存储。`StorageType` 字段用于在数据库中区分不同的存储类型。

### 上传选项

```go
type UploadOptions struct {
    ContentType string            `json:"contentType"` // 文件MIME类型
    Metadata    map[string]string `json:"metadata"`    // 元数据
    Public      bool              `json:"public"`      // 是否公开访问
}
```

### 列表选项

```go
type ListOptions struct {
    Prefix    string `json:"prefix"`    // 路径前缀
    Limit     int    `json:"limit"`     // 限制数量
    Offset    int    `json:"offset"`    // 偏移量
    SortBy    string `json:"sortBy"`    // 排序字段
    SortOrder string `json:"sortOrder"` // 排序方向 (asc/desc)
}
```

## 配置说明

### 本地存储配置

```go
type LocalConfig struct {
    BaseConfig
    RootPath string `json:"rootPath"` // 根目录路径
    BaseURL  string `json:"baseURL"`  // 基础URL（用于生成访问链接）
}

type BaseConfig struct {
    DB        *gorm.DB `json:"-"`       // gorm.DB 实例，用于存储文件信息（可选）
    TableName string   `json:"tableName"` // 存储文件信息的表名
}
```

### 数据库存储配置

```go
type DBConfig struct {
    BaseConfig
    FileTableName string `json:"fileTableName"` // 存储二进制文件的表名
}
```

**注意**: 数据库存储现在需要直接提供 `*gorm.DB` 实例，而不是通过 DSN 字符串。这样可以更好地复用数据库连接和配置。

#### 数据库存储示例

```go
import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "github.com/ffhuo/go-kits/storage"
)

// 创建数据库连接
db, err := gorm.Open(sqlite.Open("storage.db"), &gorm.Config{})
if err != nil {
    panic(err)
}

// 数据库存储配置
config := &storage.DBConfig{
    BaseConfig: storage.BaseConfig{
        DB:        db,                    // 数据库连接
        TableName: "file_metadata",       // 文件元数据表名
    },
    FileTableName: "file_storage",        // 二进制文件存储表名
}

store, err := storage.NewDatabaseStorage(config)
if err != nil {
    panic(err)
}
defer store.Close()
```

### 腾讯云COS配置

```go
type COSConfig struct {
    BaseConfig
    SecretID  string `json:"secretId"`  // 密钥ID
    SecretKey string `json:"secretKey"` // 密钥Key
    Region    string `json:"region"`    // 地域
    Bucket    string `json:"bucket"`    // 存储桶名称
    BaseURL   string `json:"baseUrl"`   // 自定义域名（可选）
}
```

## 元数据管理

所有存储方式都支持可选的数据库元数据管理。当提供 `*gorm.DB` 实例时，文件的元数据信息将被存储在数据库中，这样可以：

- 快速查询文件信息而无需访问实际存储
- 支持复杂的文件搜索和过滤
- 统一管理不同存储后端的文件元数据
- 提供更好的性能和扩展性

### 启用元数据管理

```go
import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "github.com/ffhuo/go-kits/storage"
)

// 创建数据库连接
db, err := gorm.Open(sqlite.Open("metadata.db"), &gorm.Config{})
if err != nil {
    panic(err)
}

// 本地存储 + 数据库元数据管理
config := &storage.LocalConfig{
    BaseConfig: storage.BaseConfig{
        DB:        db,                    // 提供数据库连接
        TableName: "app_file_metadata",   // 存储文件信息的表名
    },
    RootPath: "/tmp/storage",
    BaseURL:  "http://localhost:8080/files",
}

store, err := storage.NewLocalStorage(config)
```

### 不使用元数据管理

```go
// 纯本地存储（不使用数据库元数据）
config := &storage.LocalConfig{
    BaseConfig: storage.BaseConfig{
        DB:        nil, // 不提供数据库连接
        TableName: "",
    },
    RootPath: "/tmp/storage",
    BaseURL:  "http://localhost:8080/files",
}

store, err := storage.NewLocalStorage(config)
```

## 测试

运行测试：

```bash
cd storage
go test -v
```

运行示例：

```bash
cd storage/example
go run main.go
```

## 注意事项

1. **本地存储**: 确保指定的根目录有读写权限
2. **数据库存储**: 确保数据库连接正常，包会自动创建表结构
3. **COS存储**: 确保SecretID、SecretKey、Region和Bucket配置正确
4. **文件路径**: 使用Unix风格的路径分隔符（/）
5. **并发安全**: 所有存储实现都是并发安全的

## 许可证

MIT License 