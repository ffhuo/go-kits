package storage

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"strings"
	"time"

	"gorm.io/gorm"
)

// FileRecord 数据库文件记录，继承FileInfo并添加Content字段
type FileRecord struct {
	FileInfo
	Content []byte `gorm:"type:longblob" json:"-"` // 文件内容
}

// TableName 指定表名
func (FileRecord) TableName() string {
	return "file_storage"
}

// DatabaseStorage 数据库存储实现
type DatabaseStorage struct {
	db              *gorm.DB
	tableName       string
	metadataManager *MetadataManager
}

// NewDatabaseStorage 创建数据库存储实例
func NewDatabaseStorage(config *DBConfig) (*DatabaseStorage, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	tableName := config.FileTableName
	if tableName == "" {
		tableName = "file_storage"
	}

	// 创建元数据管理器（数据库存储可以选择不使用额外的元数据管理器）
	metadataManager := NewMetadataManager(config.DB, config.TableName)

	storage := &DatabaseStorage{
		db:              config.DB,
		tableName:       tableName,
		metadataManager: metadataManager,
	}

	// 自动迁移表结构
	if err := config.DB.Table(tableName).AutoMigrate(&FileRecord{}); err != nil {
		return nil, fmt.Errorf("failed to migrate table: %w", err)
	}

	return storage, nil
}

// Upload 上传文件
func (ds *DatabaseStorage) Upload(ctx context.Context, path string, reader io.Reader, opts *UploadOptions) (*FileInfo, error) {
	// 读取文件内容
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// 计算哈希值
	hash := md5.Sum(content)
	hashStr := fmt.Sprintf("%x", hash)

	now := time.Now()
	record := &FileRecord{
		FileInfo: FileInfo{
			ID:          generateFileID(path),
			Name:        getFileName(path),
			Path:        path,
			Size:        int64(len(content)),
			Hash:        hashStr,
			StorageType: string(StorageTypeDB),
			Metadata:    make(map[string]string),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Content: content,
	}

	if opts != nil {
		record.FileInfo.ContentType = opts.ContentType
		if opts.Metadata != nil {
			record.FileInfo.Metadata = opts.Metadata
		}
	}

	// 保存到数据库
	if err := ds.db.Table(ds.tableName).Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to save file to database: %w", err)
	}

	// 保存元数据到元数据管理器（如果启用）
	if ds.metadataManager != nil {
		if err := ds.metadataManager.Save(ctx, &record.FileInfo, string(StorageTypeDB)); err != nil {
			// 元数据保存失败不影响主要功能，只记录错误
			// 可以考虑添加日志记录
		}
	}

	return &record.FileInfo, nil
}

// Download 下载文件
func (ds *DatabaseStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	var record FileRecord
	if err := ds.db.Table(ds.tableName).Where("path = ?", path).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to query file: %w", err)
	}

	return io.NopCloser(strings.NewReader(string(record.Content))), nil
}

// Delete 删除文件
func (ds *DatabaseStorage) Delete(ctx context.Context, path string) error {
	result := ds.db.Table(ds.tableName).Where("path = ?", path).Delete(&FileRecord{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete file: %w", result.Error)
	}

	// 从元数据管理器中删除（如果启用）
	if ds.metadataManager != nil {
		if err := ds.metadataManager.Delete(ctx, path); err != nil {
			// 元数据删除失败不影响主要功能，只记录错误
			// 可以考虑添加日志记录
		}
	}

	return nil
}

// Exists 检查文件是否存在
func (ds *DatabaseStorage) Exists(ctx context.Context, path string) (bool, error) {
	var count int64
	if err := ds.db.Table(ds.tableName).Where("path = ?", path).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return count > 0, nil
}

// GetInfo 获取文件信息
func (ds *DatabaseStorage) GetInfo(ctx context.Context, path string) (*FileInfo, error) {
	var record FileRecord
	if err := ds.db.Table(ds.tableName).Select("id, name, path, size, content_type, hash, storage_type, created_at, updated_at, metadata").Where("path = ?", path).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to query file: %w", err)
	}

	return &record.FileInfo, nil
}

// List 列出文件
func (ds *DatabaseStorage) List(ctx context.Context, opts *ListOptions) ([]*FileInfo, error) {
	query := ds.db.Table(ds.tableName).Select("id, name, path, size, content_type, hash, storage_type, created_at, updated_at, metadata")

	if opts != nil {
		if opts.Prefix != "" {
			query = query.Where("path LIKE ?", opts.Prefix+"%")
		}

		if opts.SortBy != "" {
			order := opts.SortBy
			if opts.SortOrder == "desc" {
				order += " DESC"
			} else {
				order += " ASC"
			}
			query = query.Order(order)
		} else {
			query = query.Order("created_at DESC")
		}

		if opts.Offset > 0 {
			query = query.Offset(opts.Offset)
		}

		if opts.Limit > 0 {
			query = query.Limit(opts.Limit)
		}
	}

	var records []FileRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	files := make([]*FileInfo, len(records))
	for i := range records {
		files[i] = &records[i].FileInfo
	}

	return files, nil
}

// GetURL 获取文件访问URL
func (ds *DatabaseStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	// 数据库存储通常不直接提供URL访问，需要通过应用程序接口
	return "", fmt.Errorf("database storage does not support direct URL access")
}

// Copy 复制文件
func (ds *DatabaseStorage) Copy(ctx context.Context, srcPath, dstPath string) error {
	var srcRecord FileRecord
	if err := ds.db.Table(ds.tableName).Where("path = ?", srcPath).First(&srcRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("source file not found: %s", srcPath)
		}
		return fmt.Errorf("failed to query source file: %w", err)
	}

	now := time.Now()
	dstRecord := FileRecord{
		FileInfo: FileInfo{
			ID:          generateFileID(dstPath),
			Name:        getFileName(dstPath),
			Path:        dstPath,
			Size:        srcRecord.FileInfo.Size,
			ContentType: srcRecord.FileInfo.ContentType,
			Hash:        srcRecord.FileInfo.Hash,
			StorageType: srcRecord.FileInfo.StorageType,
			Metadata:    srcRecord.FileInfo.Metadata,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Content: srcRecord.Content,
	}

	if err := ds.db.Table(ds.tableName).Create(&dstRecord).Error; err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// 复制元数据到元数据管理器（如果启用）
	if ds.metadataManager != nil {
		if err := ds.metadataManager.Copy(ctx, srcPath, dstPath); err != nil {
			// 元数据复制失败不影响主要功能，只记录错误
			// 可以考虑添加日志记录
		}
	}

	return nil
}

// Move 移动文件
func (ds *DatabaseStorage) Move(ctx context.Context, srcPath, dstPath string) error {
	result := ds.db.Table(ds.tableName).Where("path = ?", srcPath).Updates(map[string]interface{}{
		"path":       dstPath,
		"name":       getFileName(dstPath),
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return fmt.Errorf("failed to move file: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("source file not found: %s", srcPath)
	}

	// 移动元数据管理器中的记录（如果启用）
	if ds.metadataManager != nil {
		if err := ds.metadataManager.Move(ctx, srcPath, dstPath); err != nil {
			// 元数据移动失败不影响主要功能，只记录错误
			// 可以考虑添加日志记录
		}
	}

	return nil
}

// Close 关闭存储连接
func (ds *DatabaseStorage) Close() error {
	sqlDB, err := ds.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// getFileName 从路径中提取文件名
func getFileName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}
