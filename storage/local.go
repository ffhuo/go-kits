package storage

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorage 本地存储实现
type LocalStorage struct {
	config          *LocalConfig
	metadataManager *MetadataManager
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(config *LocalConfig) (*LocalStorage, error) {
	if config.RootPath == "" {
		return nil, fmt.Errorf("rootPath is required")
	}

	// 确保根目录存在
	if err := os.MkdirAll(config.RootPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	// 创建元数据管理器
	metadataManager := NewMetadataManager(config.DB, config.TableName)

	return &LocalStorage{
		config:          config,
		metadataManager: metadataManager,
	}, nil
}

// Upload 上传文件
func (ls *LocalStorage) Upload(ctx context.Context, path string, reader io.Reader, opts *UploadOptions) (*FileInfo, error) {
	fullPath := filepath.Join(ls.config.RootPath, path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 计算哈希值并写入文件
	hash := md5.New()
	size, err := io.Copy(io.MultiWriter(file, hash), reader)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// 获取文件信息
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stat: %w", err)
	}

	fileInfo := &FileInfo{
		ID:        generateFileID(path),
		Name:      filepath.Base(path),
		Path:      path,
		Size:      size,
		Hash:      fmt.Sprintf("%x", hash.Sum(nil)),
		CreatedAt: stat.ModTime(),
		UpdatedAt: stat.ModTime(),
		Metadata:  make(map[string]string),
	}

	if opts != nil {
		fileInfo.ContentType = opts.ContentType
		if opts.Metadata != nil {
			fileInfo.Metadata = opts.Metadata
		}
	}

	// 保存文件元数据到数据库（如果启用）
	if err := ls.metadataManager.Save(ctx, fileInfo, string(StorageTypeLocal)); err != nil {
		// 元数据保存失败不影响文件上传，只记录错误
		fmt.Printf("Warning: failed to save file metadata: %v\n", err)
	}

	return fileInfo, nil
}

// Download 下载文件
func (ls *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(ls.config.RootPath, path)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete 删除文件
func (ls *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(ls.config.RootPath, path)

	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// 删除文件元数据（如果启用）
	if err := ls.metadataManager.Delete(ctx, path); err != nil {
		fmt.Printf("Warning: failed to delete file metadata: %v\n", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (ls *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(ls.config.RootPath, path)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// GetInfo 获取文件信息
func (ls *LocalStorage) GetInfo(ctx context.Context, path string) (*FileInfo, error) {
	// 优先从数据库获取元数据
	if ls.metadataManager.IsEnabled() {
		if fileInfo, err := ls.metadataManager.Get(ctx, path); err == nil {
			return fileInfo, nil
		}
	}

	// 从文件系统获取信息
	fullPath := filepath.Join(ls.config.RootPath, path)

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to get file stat: %w", err)
	}

	// 计算文件哈希值
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for hash calculation: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}

	return &FileInfo{
		ID:        generateFileID(path),
		Name:      stat.Name(),
		Path:      path,
		Size:      stat.Size(),
		Hash:      fmt.Sprintf("%x", hash.Sum(nil)),
		CreatedAt: stat.ModTime(),
		UpdatedAt: stat.ModTime(),
		Metadata:  make(map[string]string),
	}, nil
}

// List 列出文件
func (ls *LocalStorage) List(ctx context.Context, opts *ListOptions) ([]*FileInfo, error) {
	// 优先从数据库获取列表
	if ls.metadataManager.IsEnabled() {
		if files, err := ls.metadataManager.List(ctx, opts); err == nil {
			return files, nil
		}
	}

	// 从文件系统获取列表
	var files []*FileInfo

	prefix := ""
	if opts != nil && opts.Prefix != "" {
		prefix = opts.Prefix
	}

	searchPath := filepath.Join(ls.config.RootPath, prefix)

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(ls.config.RootPath, path)
		if err != nil {
			return err
		}

		// 转换为Unix风格路径
		relPath = filepath.ToSlash(relPath)

		fileInfo := &FileInfo{
			ID:        generateFileID(relPath),
			Name:      info.Name(),
			Path:      relPath,
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
			UpdatedAt: info.ModTime(),
			Metadata:  make(map[string]string),
		}

		files = append(files, fileInfo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// 应用分页
	if opts != nil {
		if opts.Offset > 0 && opts.Offset < len(files) {
			files = files[opts.Offset:]
		}
		if opts.Limit > 0 && opts.Limit < len(files) {
			files = files[:opts.Limit]
		}
	}

	return files, nil
}

// GetURL 获取文件访问URL
func (ls *LocalStorage) GetURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	if ls.config.BaseURL == "" {
		return "", fmt.Errorf("baseURL is not configured")
	}

	// 简单拼接URL，实际使用中可能需要更复杂的逻辑
	url := strings.TrimRight(ls.config.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")
	return url, nil
}

// Copy 复制文件
func (ls *LocalStorage) Copy(ctx context.Context, srcPath, dstPath string) error {
	srcFullPath := filepath.Join(ls.config.RootPath, srcPath)
	dstFullPath := filepath.Join(ls.config.RootPath, dstPath)

	// 确保目标目录存在
	dstDir := filepath.Dir(dstFullPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// 打开源文件
	srcFile, err := os.Open(srcFullPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.Create(dstFullPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// 复制文件内容
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// 复制文件元数据（如果启用）
	if err := ls.metadataManager.Copy(ctx, srcPath, dstPath); err != nil {
		fmt.Printf("Warning: failed to copy file metadata: %v\n", err)
	}

	return nil
}

// Move 移动文件
func (ls *LocalStorage) Move(ctx context.Context, srcPath, dstPath string) error {
	srcFullPath := filepath.Join(ls.config.RootPath, srcPath)
	dstFullPath := filepath.Join(ls.config.RootPath, dstPath)

	// 确保目标目录存在
	dstDir := filepath.Dir(dstFullPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// 移动文件
	if err := os.Rename(srcFullPath, dstFullPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	// 移动文件元数据（如果启用）
	if err := ls.metadataManager.Move(ctx, srcPath, dstPath); err != nil {
		fmt.Printf("Warning: failed to move file metadata: %v\n", err)
	}

	return nil
}

// Close 关闭存储连接
func (ls *LocalStorage) Close() error {
	// 本地存储无需关闭连接
	return nil
}

// generateFileID 生成文件ID
func generateFileID(path string) string {
	hash := md5.Sum([]byte(path + time.Now().String()))
	return fmt.Sprintf("%x", hash)
}
