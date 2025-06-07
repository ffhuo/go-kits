package storage

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestLocalStorage(t *testing.T) {
	// 创建临时目录
	tempDir := "/tmp/storage_test"
	defer os.RemoveAll(tempDir)

	config := &LocalConfig{
		RootPath: tempDir,
		BaseURL:  "http://localhost:8080/files",
	}

	storage, err := NewLocalStorage(config)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}
	defer storage.Close()

	testStorage(t, storage)
}

func testStorage(t *testing.T, storage Storage) {
	ctx := context.Background()

	// 测试上传文件
	content := "Hello, World!"
	reader := strings.NewReader(content)
	opts := &UploadOptions{
		ContentType: "text/plain",
		Metadata: map[string]string{
			"author": "test",
		},
	}

	fileInfo, err := storage.Upload(ctx, "test/hello.txt", reader, opts)
	if err != nil {
		t.Fatalf("Failed to upload file: %v", err)
	}

	if fileInfo.Name != "hello.txt" {
		t.Errorf("Expected file name 'hello.txt', got '%s'", fileInfo.Name)
	}

	if fileInfo.Size != int64(len(content)) {
		t.Errorf("Expected file size %d, got %d", len(content), fileInfo.Size)
	}

	// 测试文件是否存在
	exists, err := storage.Exists(ctx, "test/hello.txt")
	if err != nil {
		t.Fatalf("Failed to check file existence: %v", err)
	}

	if !exists {
		t.Error("File should exist")
	}

	// 测试获取文件信息
	info, err := storage.GetInfo(ctx, "test/hello.txt")
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	if info.Path != "test/hello.txt" {
		t.Errorf("Expected path 'test/hello.txt', got '%s'", info.Path)
	}

	// 测试下载文件
	downloadReader, err := storage.Download(ctx, "test/hello.txt")
	if err != nil {
		t.Fatalf("Failed to download file: %v", err)
	}
	defer downloadReader.Close()

	downloadedContent := make([]byte, len(content))
	_, err = downloadReader.Read(downloadedContent)
	if err != nil {
		t.Fatalf("Failed to read downloaded content: %v", err)
	}

	if string(downloadedContent) != content {
		t.Errorf("Expected content '%s', got '%s'", content, string(downloadedContent))
	}

	// 测试复制文件
	err = storage.Copy(ctx, "test/hello.txt", "test/hello_copy.txt")
	if err != nil {
		t.Fatalf("Failed to copy file: %v", err)
	}

	// 验证复制的文件存在
	exists, err = storage.Exists(ctx, "test/hello_copy.txt")
	if err != nil {
		t.Fatalf("Failed to check copied file existence: %v", err)
	}

	if !exists {
		t.Error("Copied file should exist")
	}

	// 测试移动文件
	err = storage.Move(ctx, "test/hello_copy.txt", "test/hello_moved.txt")
	if err != nil {
		t.Fatalf("Failed to move file: %v", err)
	}

	// 验证移动后的文件存在
	exists, err = storage.Exists(ctx, "test/hello_moved.txt")
	if err != nil {
		t.Fatalf("Failed to check moved file existence: %v", err)
	}

	if !exists {
		t.Error("Moved file should exist")
	}

	// 验证原文件不存在
	exists, err = storage.Exists(ctx, "test/hello_copy.txt")
	if err != nil {
		t.Fatalf("Failed to check original file existence: %v", err)
	}

	if exists {
		t.Error("Original file should not exist after move")
	}

	// 测试列出文件
	listOpts := &ListOptions{
		Prefix: "test/",
		Limit:  10,
	}

	files, err := storage.List(ctx, listOpts)
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	if len(files) < 2 {
		t.Errorf("Expected at least 2 files, got %d", len(files))
	}

	// 测试删除文件
	err = storage.Delete(ctx, "test/hello.txt")
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// 验证文件已删除
	exists, err = storage.Exists(ctx, "test/hello.txt")
	if err != nil {
		t.Fatalf("Failed to check deleted file existence: %v", err)
	}

	if exists {
		t.Error("File should not exist after deletion")
	}
}

func TestStorageManager(t *testing.T) {
	manager := NewStorageManager()

	// 创建本地存储
	tempDir := "/tmp/storage_manager_test"
	defer os.RemoveAll(tempDir)

	config := &LocalConfig{
		RootPath: tempDir,
		BaseURL:  "http://localhost:8080/files",
	}

	localStorage, err := NewLocalStorage(config)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	// 注册存储
	manager.Register("local", localStorage)

	// 获取存储
	storage, err := manager.Get("local")
	if err != nil {
		t.Fatalf("Failed to get storage: %v", err)
	}

	if storage == nil {
		t.Error("Storage should not be nil")
	}

	// 测试不存在的存储
	_, err = manager.Get("nonexistent")
	if err == nil {
		t.Error("Should return error for nonexistent storage")
	}

	// 测试列出存储
	names := manager.List()
	if len(names) != 1 || names[0] != "local" {
		t.Errorf("Expected ['local'], got %v", names)
	}

	// 关闭管理器
	err = manager.Close()
	if err != nil {
		t.Fatalf("Failed to close manager: %v", err)
	}
}

func TestNewStorage(t *testing.T) {
	// 测试本地存储配置
	config := &Config{
		Type: StorageTypeLocal,
		Settings: map[string]interface{}{
			"rootPath": "/tmp/test_storage",
			"baseURL":  "http://localhost:8080/files",
		},
	}

	storage, err := NewStorage(config)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if storage == nil {
		t.Error("Storage should not be nil")
	}

	// 测试无效配置
	invalidConfig := &Config{
		Type:     "invalid",
		Settings: map[string]interface{}{},
	}

	_, err = NewStorage(invalidConfig)
	if err == nil {
		t.Error("Should return error for invalid storage type")
	}
}
