package prometheus

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.uber.org/zap"
)

var (
	RefreshInterval     = 15 * time.Second
	defaultHttpHistoram = "requests_cost"
	defaultPanicCounter = "panic_total"
)

type Prometheus struct {
	cs  sync.Map
	log *zap.Logger
}

type Config func(*Prometheus)

func Logger(log *zap.Logger) Config {
	return func(p *Prometheus) {
		p.log = log
	}
}

func Handler(engine *gin.Engine) Config {
	return func(p *Prometheus) {
		engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}
}

// New new gin prometheus
func Init(configs ...Config) *Prometheus {
	var (
		// httpHistogram prometheus 模型
		httpHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   "http_server",
			Subsystem:   "",
			Name:        defaultHttpHistoram,
			Help:        "Histogram of response latency (seconds) of http handlers.",
			ConstLabels: nil,
			Buckets:     nil,
		}, []string{"method", "code", "uri"})

		panicCounter = prometheus.NewCounter(prometheus.CounterOpts{
			Name: defaultPanicCounter,
			Help: "The total number of server panic.",
		})
	)

	gp := &Prometheus{}
	for _, conf := range configs {
		conf(gp)
	}

	gp.cs.Store(defaultHttpHistoram, httpHistogram)
	gp.cs.Store(defaultPanicCounter, panicCounter)

	prometheus.MustRegister(httpHistogram, panicCounter)
	return gp
}

func (gp *Prometheus) PanicInc() {
	panicCounter, ok := gp.cs.Load(defaultPanicCounter)
	if !ok {
		return
	}
	panicCounter.(prometheus.Counter).Inc()
}

func (gp *Prometheus) Push(addr, job string) {
	pusher := push.New(addr, job)

	for range time.Tick(RefreshInterval) {
		err := pusher.Push()
		if err != nil && gp.log != nil {
			gp.log.Error("prometheus push err:", zap.Error(err))
		}
	}
}

func (gp *Prometheus) RegisterCollector(name string, cs prometheus.Collector) error {
	_, ok := gp.cs.Load(name)
	if ok {
		return errors.New("name exist")
	}
	gp.cs.Store(name, cs)
	return nil
}

func (gp *Prometheus) LoadCollector(name string) (interface{}, bool) {
	v, ok := gp.cs.Load(name)
	if !ok {
		return nil, false
	}
	return v, true
}

// Middleware set gin middleware
func (gp *Prometheus) Middleware(ignored map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 过滤请求
		if ignored[c.Request.URL.String()] {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		httpHistogram, ok := gp.cs.Load(defaultHttpHistoram)
		if !ok {
			c.Next()
			return
		}

		(httpHistogram.(*prometheus.HistogramVec)).WithLabelValues(
			c.Request.Method,
			strconv.Itoa(c.Writer.Status()),
			c.Request.URL.Path,
		).Observe(time.Since(start).Seconds())
	}
}
