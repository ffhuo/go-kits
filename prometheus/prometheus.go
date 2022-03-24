package prometheus

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// httpHistogram prometheus 模型
	httpHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "http_server",
		Subsystem:   "",
		Name:        "requests_seconds",
		Help:        "Histogram of response latency (seconds) of http handlers.",
		ConstLabels: nil,
		Buckets:     nil,
	}, []string{"method", "code", "uri"})
)

type Prometheus struct {
	engine  *gin.Engine
	ignored map[string]bool
}

type Config func(*Prometheus)

// Ignore 添加忽略的路径
func Ignore(path map[string]bool) Config {
	return func(gp *Prometheus) {
		gp.ignored = path
	}
}

// New new gin prometheus
func Init(e *gin.Engine, options ...Config) *Prometheus {
	if e == nil {
		return nil
	}

	gp := &Prometheus{
		engine:  e,
		ignored: map[string]bool{},
	}

	for _, o := range options {
		o(gp)
	}

	e.GET("/metrics", gin.WrapH(promhttp.Handler())) // register prometheus
	prometheus.MustRegister(httpHistogram)
	return gp
}

// Middleware set gin middleware
func (gp *Prometheus) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 过滤请求
		if gp.ignored[c.Request.URL.String()] {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		httpHistogram.WithLabelValues(
			c.Request.Method,
			strconv.Itoa(c.Writer.Status()),
			c.Request.URL.Path,
		).Observe(time.Since(start).Seconds())
	}
}
