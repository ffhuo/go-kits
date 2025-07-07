package main

import (
	"context"
	"time"

	"github.com/ffhuo/go-kits/ginmiddleware"
	"github.com/ffhuo/go-kits/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	// 创建日志实例
	log, err := logger.New(
		logger.WithInfoLevel(),
		logger.WithFormatJSON(),
		logger.WithFileRotationP("logs/app.log", 100, 7, 30),
	)
	if err != nil {
		panic(err)
	}

	// 创建Gin实例
	r := gin.New()

	// 添加中间件（按推荐顺序）
	r.Use(ginmiddleware.RequestID())
	r.Use(ginmiddleware.CORS(&ginmiddleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
	}))
	r.Use(ginmiddleware.LoggerMiddleware(&ginmiddleware.LoggerConfig{
		Logger:        log,
		SkipPaths:     []string{"/health"},
		SlowThreshold: 200 * time.Millisecond,
	}))
	r.Use(ginmiddleware.RecoveryWithLogger(log))

	// 路由定义
	r.GET("/", func(c *gin.Context) {
		requestID := ginmiddleware.GetRequestID(c)
		c.JSON(200, gin.H{
			"message":    "Hello World",
			"request_id": requestID,
			"timestamp":  time.Now(),
		})
	})

	r.GET("/slow", func(c *gin.Context) {
		// 模拟慢请求
		time.Sleep(300 * time.Millisecond)
		c.JSON(200, gin.H{
			"message": "This is a slow request",
		})
	})

	r.GET("/panic", func(c *gin.Context) {
		// 模拟panic
		panic("这是一个测试panic")
	})

	r.GET("/error", func(c *gin.Context) {
		// 模拟错误响应
		c.JSON(400, gin.H{
			"error": "Bad Request",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		// 健康检查接口（不会被日志记录）
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/data", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		requestID := ginmiddleware.GetRequestID(c)
		c.JSON(200, gin.H{
			"message":    "Data received",
			"request_id": requestID,
			"data":       data,
		})
	})

	// 启动服务器
	log.Info(context.Background(), "Starting server on :8080")
	r.Run(":8080")
}
