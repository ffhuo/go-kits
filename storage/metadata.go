package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// 使用统一的FileInfo结构体，不再需要单独的FileMetadata

// MetadataManager 文件元数据管理器
type MetadataManager struct {
	db        *gorm.DB
	tableName string
	enabled   bool
}

// NewMetadataManager 创建元数据管理器
func NewMetadataManager(db *gorm.DB, tableName string) *MetadataManager {
	if db == nil {
		return &MetadataManager{enabled: false}
	}

	if tableName == "" {
		tableName = "file_metadata"
	}

	manager := &MetadataManager{
		db:        db,
		tableName: tableName,
		enabled:   true,
	}

	// 自动迁移表结构
	if err := db.Table(tableName).AutoMigrate(&FileInfo{}); err != nil {
		// 如果迁移失败，禁用元数据管理
		manager.enabled = false
	}

	return manager
}

// IsEnabled 检查是否启用了元数据管理
func (mm *MetadataManager) IsEnabled() bool {
	return mm.enabled
}

// Save 保存文件元数据
func (mm *MetadataManager) Save(ctx context.Context, fileInfo *FileInfo, storageType string) error {
	if !mm.enabled {
		return nil
	}

	// 设置存储类型
	fileInfo.StorageType = storageType

	return mm.db.Table(mm.tableName).Create(fileInfo).Error
}

// Update 更新文件元数据
func (mm *MetadataManager) Update(ctx context.Context, path string, updates map[string]interface{}) error {
	if !mm.enabled {
		return nil
	}

	updates["updated_at"] = time.Now()
	return mm.db.Table(mm.tableName).Where("path = ?", path).Updates(updates).Error
}

// Get 获取文件元数据
func (mm *MetadataManager) Get(ctx context.Context, path string) (*FileInfo, error) {
	if !mm.enabled {
		return nil, fmt.Errorf("metadata manager is not enabled")
	}

	var fileInfo FileInfo
	if err := mm.db.Table(mm.tableName).Where("path = ?", path).First(&fileInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, err
	}

	return &fileInfo, nil
}

// Delete 删除文件元数据
func (mm *MetadataManager) Delete(ctx context.Context, path string) error {
	if !mm.enabled {
		return nil
	}

	return mm.db.Table(mm.tableName).Where("path = ?", path).Delete(&FileInfo{}).Error
}

// Exists 检查文件元数据是否存在
func (mm *MetadataManager) Exists(ctx context.Context, path string) (bool, error) {
	if !mm.enabled {
		return false, nil
	}

	var count int64
	if err := mm.db.Table(mm.tableName).Where("path = ?", path).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// List 列出文件元数据
func (mm *MetadataManager) List(ctx context.Context, opts *ListOptions) ([]*FileInfo, error) {
	if !mm.enabled {
		return nil, fmt.Errorf("metadata manager is not enabled")
	}

	query := mm.db.Table(mm.tableName)

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

	var fileInfoList []FileInfo
	if err := query.Find(&fileInfoList).Error; err != nil {
		return nil, err
	}

	files := make([]*FileInfo, len(fileInfoList))
	for i := range fileInfoList {
		files[i] = &fileInfoList[i]
	}

	return files, nil
}

// Copy 复制文件元数据
func (mm *MetadataManager) Copy(ctx context.Context, srcPath, dstPath string) error {
	if !mm.enabled {
		return nil
	}

	var srcFileInfo FileInfo
	if err := mm.db.Table(mm.tableName).Where("path = ?", srcPath).First(&srcFileInfo).Error; err != nil {
		return err
	}

	now := time.Now()
	dstFileInfo := FileInfo{
		ID:          generateFileID(dstPath),
		Name:        getFileName(dstPath),
		Path:        dstPath,
		Size:        srcFileInfo.Size,
		ContentType: srcFileInfo.ContentType,
		Hash:        srcFileInfo.Hash,
		StorageType: srcFileInfo.StorageType,
		Metadata:    srcFileInfo.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return mm.db.Table(mm.tableName).Create(&dstFileInfo).Error
}

// Move 移动文件元数据
func (mm *MetadataManager) Move(ctx context.Context, srcPath, dstPath string) error {
	if !mm.enabled {
		return nil
	}

	updates := map[string]interface{}{
		"path":       dstPath,
		"name":       getFileName(dstPath),
		"updated_at": time.Now(),
	}

	return mm.db.Table(mm.tableName).Where("path = ?", srcPath).Updates(updates).Error
}
