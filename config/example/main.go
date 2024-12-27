package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ffhuo/go-kits/config"
	"github.com/fsnotify/fsnotify"
)

func main() {
	// 创建配置实例
	cfg := config.New().
		AddPath("configs").                // 添加配置文件路径
		SetEnvPrefix("APP").              // 设置环境变量前缀
		AutomaticEnv()                    // 启用自动环境变量绑定

	// 加载主配置文件
	if err := cfg.LoadFile("config.yaml"); err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}

	// 合并其他配置文件（可选）
	if err := cfg.MergeFile("config.local.yaml"); err != nil {
		log.Printf("No local config file found: %v", err)
	}

	// 监听配置文件变更
	cfg.Watch(func(event fsnotify.Event) {
		if event.Op&fsnotify.Write == fsnotify.Write {
			log.Printf("Config file changed: %s", event.Name)
			// 在这里处理配置更新
			reloadConfig(cfg)
		}
	})

	// 使用配置值
	var mysqlConfig config.MySQLConfig
	if err := cfg.Unmarshal("mysql", &mysqlConfig); err != nil {
		log.Fatalf("Failed to unmarshal MySQL config: %v", err)
	}

	var redisConfig config.RedisConfig
	if err := cfg.Unmarshal("redis", &redisConfig); err != nil {
		log.Fatalf("Failed to unmarshal Redis config: %v", err)
	}

	var mongoConfig config.MongoDBConfig
	if err := cfg.Unmarshal("mongodb", &mongoConfig); err != nil {
		log.Fatalf("Failed to unmarshal MongoDB config: %v", err)
	}

	// 打印配置信息
	fmt.Printf("MySQL Config: %+v\n", mysqlConfig)
	fmt.Printf("Redis Config: %+v\n", redisConfig)
	fmt.Printf("MongoDB Config: %+v\n", mongoConfig)

	// 保持程序运行以观察配置变更
	for {
		time.Sleep(time.Second)
	}
}

func reloadConfig(cfg *config.Config) {
	var mysqlConfig config.MySQLConfig
	if err := cfg.Unmarshal("mysql", &mysqlConfig); err != nil {
		log.Printf("Failed to reload MySQL config: %v", err)
		return
	}

	// 在这里重新配置 MySQL 连接
	fmt.Printf("MySQL Config updated: %+v\n", mysqlConfig)
}
