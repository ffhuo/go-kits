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

func databaseStorageExample() {
	ctx := context.Background()

	// 注意：这里需要实际的数据库连接
	// 为了演示，我们注释掉数据库相关代码
	/*
		// 创建数据库连接
		db, err := gorm.Open(sqlite.Open("/tmp/database_storage.db"), &gorm.Config{})
		if err != nil {
			log.Fatalf("连接数据库失败: %v", err)
		}
	*/

	// 数据库存储配置
	config := &storage.DBConfig{
		BaseConfig: storage.BaseConfig{
			DB:        nil,             // db, // 传入*gorm.DB实例
			TableName: "file_metadata", // 文件元数据表名
		},
		FileTableName: "file_storage", // 二进制文件存储表名
	}

	// 创建存储实例
	store, err := storage.NewDatabaseStorage(config)
	if err != nil {
		log.Fatalf("创建数据库存储失败: %v", err)
	}
	defer store.Close()

	// 上传文件
	content := "这是存储在数据库中的文件内容"
	reader := strings.NewReader(content)
	opts := &storage.UploadOptions{
		ContentType: "text/plain; charset=utf-8",
		Metadata: map[string]string{
			"author":      "李四",
			"department":  "研发部",
			"project":     "数据库存储系统",
			"version":     "2.0",
			"description": "演示数据库存储功能",
		},
	}

	fileInfo, err := store.Upload(ctx, "database/test.txt", reader, opts)
	if err != nil {
		log.Fatalf("上传文件失败: %v", err)
	}

	fmt.Printf("文件上传到数据库成功:\n")
	fmt.Printf("  ID: %s\n", fileInfo.ID)
	fmt.Printf("  路径: %s\n", fileInfo.Path)
	fmt.Printf("  大小: %d 字节\n", fileInfo.Size)
	fmt.Printf("  存储类型: %s\n", fileInfo.StorageType)
	fmt.Printf("  元数据: %+v\n", fileInfo.Metadata)

	// 检查文件是否存在
	exists, err := store.Exists(ctx, "database/test.txt")
	if err != nil {
		log.Fatalf("检查文件存在性失败: %v", err)
	}
	fmt.Printf("文件存在: %t\n", exists)

	// 获取文件信息
	info, err := store.GetInfo(ctx, "database/test.txt")
	if err != nil {
		log.Fatalf("获取文件信息失败: %v", err)
	}
	fmt.Printf("文件信息: %s (%d 字节)\n", info.Path, info.Size)

	// 下载文件
	reader2, err := store.Download(ctx, "database/test.txt")
	if err != nil {
		log.Fatalf("下载文件失败: %v", err)
	}
	defer reader2.Close()

	// 读取下载的内容
	downloadedContent := make([]byte, info.Size)
	_, err = reader2.Read(downloadedContent)
	if err != nil {
		log.Fatalf("读取下载内容失败: %v", err)
	}
	fmt.Printf("下载的文件内容: %s\n", string(downloadedContent))

	// 列出文件
	listOpts := &storage.ListOptions{
		Prefix: "database/",
		Limit:  10,
	}

	files, err := store.List(ctx, listOpts)
	if err != nil {
		log.Fatalf("列出文件失败: %v", err)
	}

	fmt.Printf("找到 %d 个文件:\n", len(files))
	for _, file := range files {
		fmt.Printf("  - %s (作者: %s, 项目: %s)\n",
			file.Path,
			file.Metadata["author"],
			file.Metadata["project"])
	}

	// 复制文件
	err = store.Copy(ctx, "database/test.txt", "database/test_copy.txt")
	if err != nil {
		log.Fatalf("复制文件失败: %v", err)
	}
	fmt.Println("文件复制成功: database/test.txt -> database/test_copy.txt")

	// 移动文件
	err = store.Move(ctx, "database/test_copy.txt", "database/moved_test.txt")
	if err != nil {
		log.Fatalf("移动文件失败: %v", err)
	}
	fmt.Println("文件移动成功: database/test_copy.txt -> database/moved_test.txt")

	// 删除文件
	err = store.Delete(ctx, "database/moved_test.txt")
	if err != nil {
		log.Fatalf("删除文件失败: %v", err)
	}
	fmt.Println("文件删除成功: database/moved_test.txt")
}

func runDatabaseExample() {
	fmt.Println("=== 数据库存储示例 ===")
	databaseStorageExample()
}
