package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ffhuo/go-kits/storage"
)

func main() {
	ctx := context.Background()

	// 示例1: 使用本地存储
	fmt.Println("=== 本地存储示例 ===")
	localExample(ctx)

	// 示例2: 使用存储管理器
	fmt.Println("\n=== 存储管理器示例 ===")
	managerExample(ctx)

	// 示例3: 使用工厂函数
	fmt.Println("\n=== 工厂函数示例 ===")
	factoryExample(ctx)
}

func localExample(ctx context.Context) {
	// 创建本地存储配置
	config := &storage.LocalConfig{
		RootPath: "/tmp/storage_example",
		BaseURL:  "http://localhost:8080/files",
	}

	// 创建本地存储实例
	localStorage, err := storage.NewLocalStorage(config)
	if err != nil {
		log.Fatalf("创建本地存储失败: %v", err)
	}
	defer localStorage.Close()

	// 上传文件
	content := "这是一个测试文件的内容"
	reader := strings.NewReader(content)
	opts := &storage.UploadOptions{
		ContentType: "text/plain; charset=utf-8",
		Metadata: map[string]string{
			"author":      "张三",
			"description": "测试文件",
		},
	}

	fileInfo, err := localStorage.Upload(ctx, "documents/test.txt", reader, opts)
	if err != nil {
		log.Fatalf("上传文件失败: %v", err)
	}

	fmt.Printf("文件上传成功:\n")
	fmt.Printf("  ID: %s\n", fileInfo.ID)
	fmt.Printf("  名称: %s\n", fileInfo.Name)
	fmt.Printf("  路径: %s\n", fileInfo.Path)
	fmt.Printf("  大小: %d 字节\n", fileInfo.Size)
	fmt.Printf("  哈希: %s\n", fileInfo.Hash)
	fmt.Printf("  创建时间: %s\n", fileInfo.CreatedAt.Format("2006-01-02 15:04:05"))

	// 检查文件是否存在
	exists, err := localStorage.Exists(ctx, "documents/test.txt")
	if err != nil {
		log.Fatalf("检查文件存在性失败: %v", err)
	}
	fmt.Printf("文件存在: %t\n", exists)

	// 下载文件
	downloadReader, err := localStorage.Download(ctx, "documents/test.txt")
	if err != nil {
		log.Fatalf("下载文件失败: %v", err)
	}
	defer downloadReader.Close()

	downloadedContent := make([]byte, fileInfo.Size)
	_, err = downloadReader.Read(downloadedContent)
	if err != nil {
		log.Fatalf("读取下载内容失败: %v", err)
	}
	fmt.Printf("下载的文件内容: %s\n", string(downloadedContent))

	// 获取文件访问URL
	url, err := localStorage.GetURL(ctx, "documents/test.txt", 0)
	if err != nil {
		log.Printf("获取文件URL失败: %v", err)
	} else {
		fmt.Printf("文件访问URL: %s\n", url)
	}

	// 列出文件
	listOpts := &storage.ListOptions{
		Prefix: "documents/",
		Limit:  10,
	}
	files, err := localStorage.List(ctx, listOpts)
	if err != nil {
		log.Fatalf("列出文件失败: %v", err)
	}

	fmt.Printf("找到 %d 个文件:\n", len(files))
	for _, file := range files {
		fmt.Printf("  - %s (%d 字节)\n", file.Path, file.Size)
	}
}

func managerExample(ctx context.Context) {
	// 创建存储管理器
	manager := storage.NewStorageManager()
	defer manager.Close()

	// 创建并注册本地存储
	localConfig := &storage.LocalConfig{
		RootPath: "/tmp/storage_manager_example",
		BaseURL:  "http://localhost:8080/files",
	}
	localStorage, err := storage.NewLocalStorage(localConfig)
	if err != nil {
		log.Fatalf("创建本地存储失败: %v", err)
	}
	manager.Register("local", localStorage)

	// 获取存储实例
	store, err := manager.Get("local")
	if err != nil {
		log.Fatalf("获取存储实例失败: %v", err)
	}

	// 使用存储实例
	content := "通过管理器使用存储"
	reader := strings.NewReader(content)
	fileInfo, err := store.Upload(ctx, "manager/test.txt", reader, nil)
	if err != nil {
		log.Fatalf("上传文件失败: %v", err)
	}

	fmt.Printf("通过管理器上传文件成功: %s\n", fileInfo.Path)

	// 列出所有注册的存储
	storageNames := manager.List()
	fmt.Printf("已注册的存储: %v\n", storageNames)
}

func factoryExample(ctx context.Context) {
	// 使用工厂函数创建本地存储
	config := &storage.Config{
		Type: storage.StorageTypeLocal,
		Settings: map[string]interface{}{
			"rootPath": "/tmp/factory_example",
			"baseURL":  "http://localhost:8080/files",
		},
	}

	store, err := storage.NewStorage(config)
	if err != nil {
		log.Fatalf("创建存储失败: %v", err)
	}
	defer store.Close()

	// 上传文件
	content := "通过工厂函数创建的存储"
	reader := strings.NewReader(content)
	fileInfo, err := store.Upload(ctx, "factory/test.txt", reader, nil)
	if err != nil {
		log.Fatalf("上传文件失败: %v", err)
	}

	fmt.Printf("通过工厂函数上传文件成功: %s\n", fileInfo.Path)

	// 演示数据库存储配置（注释掉，因为需要实际的数据库连接）
	/*
		dbConfig := &storage.Config{
			Type: storage.StorageTypeDB,
			Settings: map[string]interface{}{
				"dsn":       "user:password@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True&loc=Local",
				"tableName": "file_storage",
			},
		}
	*/

	// 演示COS存储配置（注释掉，因为需要实际的COS凭证）
	/*
		cosConfig := &storage.Config{
			Type: storage.StorageTypeCOS,
			Settings: map[string]interface{}{
				"secretId":  "your-secret-id",
				"secretKey": "your-secret-key",
				"region":    "ap-beijing",
				"bucket":    "your-bucket-name",
				"baseUrl":   "https://your-custom-domain.com", // 可选
			},
		}
	*/

	fmt.Println("工厂函数示例完成")
}
