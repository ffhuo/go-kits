# mysql监控

```golang
package main

import (
	"context"

	"github.com/ffhuo/go-kits/db"
	"github.com/ffhuo/go-kits/logger"
	"gorm.io/plugin/prometheus"
)

func main() {
	log, err := logger.New(logger.WithTimeLayout("2006-01-02 15:04:05"))
	if err != nil {
		panic(err)
	}
	log.Info(context.Background(), "hello world")

	collector := &db.Metric{}
	db.New(db.SetWriteAddr("xxxx"), db.Logger(log), db.Plugins(
		db.NewTrace(collector),
		prometheus.New(prometheus.Config{
			DBName:          "db1",                         // 使用 `DBName` 作为指标 label
			RefreshInterval: 15,                            // 指标刷新频率（默认为 15 秒）
			PushAddr:        "http://xxx.xxx.xxx.xxx:9091", // 如果配置了 `PushAddr`，则推送指标
			StartServer:     false,                         // 启用一个 http 服务来暴露指标
			HTTPServerPort:  8080,                          // 配置 http 服务监听端口，默认端口为 8080 （如果您配置了多个，只有第一个 `HTTPServerPort` 会被使用）
			MetricsCollector: []prometheus.MetricsCollector{
				collector,
			}, // 用户自定义指标
		}),
	))
}
```