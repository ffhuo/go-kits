package storage

import (
	"context"
	"io"
	"time"

	"gorm.io/gorm"
)

// FileInfo 文件信息结构体
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

// UploadOptions 上传选项
type UploadOptions struct {
	ContentType string            `json:"contentType"` // 文件MIME类型
	Metadata    map[string]string `json:"metadata"`    // 元数据
	Public      bool              `json:"public"`      // 是否公开访问
}

// ListOptions 列表查询选项
type ListOptions struct {
	Prefix    string `json:"prefix"`    // 路径前缀
	Limit     int    `json:"limit"`     // 限制数量
	Offset    int    `json:"offset"`    // 偏移量
	SortBy    string `json:"sortBy"`    // 排序字段
	SortOrder string `json:"sortOrder"` // 排序方向 (asc/desc)
}

// Storage 存储接口
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

// StorageType 存储类型
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeDB    StorageType = "database"
	StorageTypeCOS   StorageType = "cos"
)

// Config 存储配置
type Config struct {
	Type     StorageType            `json:"type"`     // 存储类型
	Settings map[string]interface{} `json:"settings"` // 配置参数
}

// BaseConfig 基础配置，包含可选的数据库存储
type BaseConfig struct {
	DB        *gorm.DB `json:"-"`         // gorm.DB 实例，用于存储文件信息
	TableName string   `json:"tableName"` // 存储文件信息的表名
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	BaseConfig
	RootPath string `json:"rootPath"` // 根目录路径
	BaseURL  string `json:"baseURL"`  // 基础URL（用于生成访问链接）
}

// DBConfig 数据库存储配置
type DBConfig struct {
	BaseConfig
	FileTableName string `json:"fileTableName"` // 存储二进制文件的表名
}

// COSConfig 腾讯云COS配置
type COSConfig struct {
	BaseConfig
	SecretID  string `json:"secretId"`  // 密钥ID
	SecretKey string `json:"secretKey"` // 密钥Key
	Region    string `json:"region"`    // 地域
	Bucket    string `json:"bucket"`    // 存储桶名称
	BaseURL   string `json:"baseUrl"`   // 自定义域名（可选）
}
