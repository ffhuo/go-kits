package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ffhuo/go-kits/storage"
	// "gorm.io/driver/sqlite"
	// "gorm.io/gorm"
)

func runMetadataExample() {
	ctx := context.Background()

	// 示例1: 本地存储 + 数据库元数据管理
	fmt.Println("=== 本地存储 + 数据库元数据管理示例 ===")
	localWithMetadataExample(ctx)

	// 示例2: 纯本地存储（无数据库元数据）
	fmt.Println("\n=== 纯本地存储示例 ===")
	pureLocalExample(ctx)
}

func localWithMetadataExample(ctx context.Context) {
	// 注意：这里需要实际的数据库连接
	// 为了演示，我们注释掉数据库相关代码
	/*
		// 创建数据库连接
		db, err := gorm.Open(sqlite.Open("/tmp/storage_metadata.db"), &gorm.Config{})
		if err != nil {
			log.Fatalf("连接数据库失败: %v", err)
		}
	*/

	// 创建本地存储配置（带数据库元数据管理）
	config := &storage.LocalConfig{
		BaseConfig: storage.BaseConfig{
			DB:        nil,                 // db, // 传入*gorm.DB实例
			TableName: "app_file_metadata", // 存储文件信息的表名
		},
		RootPath: "/tmp/storage_with_metadata",
		BaseURL:  "http://localhost:8080/files",
	}

	// 创建存储实例
	store, err := storage.NewLocalStorage(config)
	if err != nil {
		log.Fatalf("创建存储失败: %v", err)
	}
	defer store.Close()

	// 上传文件
	content := "这是一个带有元数据管理的文件"
	reader := strings.NewReader(content)
	opts := &storage.UploadOptions{
		ContentType: "text/plain; charset=utf-8",
		Metadata: map[string]string{
			"author":     "张三",
			"department": "技术部",
			"project":    "文件管理系统",
			"version":    "1.0",
		},
	}

	fileInfo, err := store.Upload(ctx, "projects/document.txt", reader, opts)
	if err != nil {
		log.Fatalf("上传文件失败: %v", err)
	}

	fmt.Printf("文件上传成功:\n")
	fmt.Printf("  ID: %s\n", fileInfo.ID)
	fmt.Printf("  路径: %s\n", fileInfo.Path)
	fmt.Printf("  大小: %d 字节\n", fileInfo.Size)
	fmt.Printf("  元数据: %+v\n", fileInfo.Metadata)

	// 获取文件信息（会优先从数据库获取）
	info, err := store.GetInfo(ctx, "projects/document.txt")
	if err != nil {
		log.Fatalf("获取文件信息失败: %v", err)
	}

	fmt.Printf("从存储获取的文件信息:\n")
	fmt.Printf("  路径: %s\n", info.Path)
	fmt.Printf("  元数据: %+v\n", info.Metadata)

	// 列出文件（会优先从数据库获取）
	listOpts := &storage.ListOptions{
		Prefix: "projects/",
		Limit:  10,
	}

	files, err := store.List(ctx, listOpts)
	if err != nil {
		log.Fatalf("列出文件失败: %v", err)
	}

	fmt.Printf("找到 %d 个文件:\n", len(files))
	for _, file := range files {
		fmt.Printf("  - %s (作者: %s)\n", file.Path, file.Metadata["author"])
	}
}

func pureLocalExample(ctx context.Context) {
	// 创建本地存储配置（不使用数据库元数据管理）
	config := &storage.LocalConfig{
		BaseConfig: storage.BaseConfig{
			DB:        nil, // 不提供数据库连接
			TableName: "",
		},
		RootPath: "/tmp/storage_pure_local",
		BaseURL:  "http://localhost:8080/files",
	}

	// 创建存储实例
	store, err := storage.NewLocalStorage(config)
	if err != nil {
		log.Fatalf("创建存储失败: %v", err)
	}
	defer store.Close()

	// 上传文件
	content := "这是一个纯本地存储的文件"
	reader := strings.NewReader(content)
	opts := &storage.UploadOptions{
		ContentType: "text/plain; charset=utf-8",
		Metadata: map[string]string{
			"note": "这些元数据不会保存到数据库",
		},
	}

	fileInfo, err := store.Upload(ctx, "simple/test.txt", reader, opts)
	if err != nil {
		log.Fatalf("上传文件失败: %v", err)
	}

	fmt.Printf("纯本地存储文件上传成功: %s\n", fileInfo.Path)

	// 获取文件信息（直接从文件系统获取）
	info, err := store.GetInfo(ctx, "simple/test.txt")
	if err != nil {
		log.Fatalf("获取文件信息失败: %v", err)
	}

	fmt.Printf("文件信息: %s (%d 字节)\n", info.Path, info.Size)
}

// 使用工厂函数创建带元数据管理的存储
func factoryWithMetadataExample() {
	// 注意：这里需要实际的数据库连接
	/*
		db, err := gorm.Open(sqlite.Open("/tmp/factory_metadata.db"), &gorm.Config{})
		if err != nil {
			log.Fatalf("连接数据库失败: %v", err)
		}
	*/

	config := &storage.Config{
		Type: storage.StorageTypeLocal,
		Settings: map[string]interface{}{
			"rootPath":    "/tmp/factory_storage",
			"baseURL":     "http://localhost:8080/files",
			"tablePrefix": "factory",
			// "db": db, // 需要特殊处理，因为map[string]interface{}不能直接存储*gorm.DB
		},
	}

	store, err := storage.NewStorage(config)
	if err != nil {
		log.Fatalf("创建存储失败: %v", err)
	}
	defer store.Close()

	fmt.Println("工厂函数创建的存储实例已就绪")
}
