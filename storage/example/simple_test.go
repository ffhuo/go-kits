package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/ffhuo/go-kits/storage"
)

func simpleTest() {
	ctx := context.Background()

	// 测试纯本地存储（不使用数据库元数据）
	config := &storage.LocalConfig{
		BaseConfig: storage.BaseConfig{
			DB:        nil, // 不提供数据库连接
			TableName: "",
		},
		RootPath: "/tmp/storage_simple_test",
		BaseURL:  "http://localhost:8080/files",
	}

	// 创建存储实例
	store, err := storage.NewLocalStorage(config)
	if err != nil {
		fmt.Printf("创建存储失败: %v\n", err)
		return
	}
	defer store.Close()

	// 上传文件
	content := "这是一个测试文件"
	reader := strings.NewReader(content)
	opts := &storage.UploadOptions{
		ContentType: "text/plain; charset=utf-8",
		Metadata: map[string]string{
			"test": "simple",
		},
	}

	fileInfo, err := store.Upload(ctx, "test/simple.txt", reader, opts)
	if err != nil {
		fmt.Printf("上传文件失败: %v\n", err)
		return
	}

	fmt.Printf("文件上传成功: %s (%d 字节)\n", fileInfo.Path, fileInfo.Size)

	// 检查文件是否存在
	exists, err := store.Exists(ctx, "test/simple.txt")
	if err != nil {
		fmt.Printf("检查文件存在性失败: %v\n", err)
		return
	}

	fmt.Printf("文件存在: %t\n", exists)

	// 获取文件信息
	info, err := store.GetInfo(ctx, "test/simple.txt")
	if err != nil {
		fmt.Printf("获取文件信息失败: %v\n", err)
		return
	}

	fmt.Printf("文件信息: %s, 存储类型: %s\n", info.Path, info.StorageType)
}

func init() {
	fmt.Println("=== 简单测试 ===")
	simpleTest()
}
