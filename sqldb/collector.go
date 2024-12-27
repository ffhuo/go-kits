package sqldb

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	sqlDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sql_duration_seconds",
			Help:    "SQL execution duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method"},
	)

	sqlErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sql_errors_total",
			Help: "Total number of SQL errors",
		},
		[]string{"method"},
	)

	sqlRows = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sql_rows",
			Help:    "Number of rows affected by SQL operations",
			Buckets: []float64{0, 1, 10, 100, 1000, 10000},
		},
		[]string{"method"},
	)

	sqlSlowQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sql_slow_queries_total",
			Help: "Total number of slow SQL queries",
		},
		[]string{"method"},
	)
)

func init() {
	prometheus.MustRegister(sqlDuration)
	prometheus.MustRegister(sqlErrors)
	prometheus.MustRegister(sqlRows)
	prometheus.MustRegister(sqlSlowQueries)
}

// SQLMetrics SQL 指标收集器
type SQLMetrics struct {
	slowThreshold time.Duration
}

// NewSQLMetrics 创建 SQL 指标收集器
func NewSQLMetrics(slowThreshold time.Duration) *SQLMetrics {
	return &SQLMetrics{
		slowThreshold: slowThreshold,
	}
}

// Collect 收集 SQL 执行指标
func (m *SQLMetrics) Collect(trace *SQLTrace) {
	// 记录执行时间
	sqlDuration.WithLabelValues(trace.Method).Observe(trace.CostSeconds)

	// 记录影响行数
	if trace.Rows > 0 {
		sqlRows.WithLabelValues(trace.Method).Observe(float64(trace.Rows))
	}

	// 检查是否是慢查询
	duration := time.Duration(trace.CostSeconds * float64(time.Second))
	if duration > m.slowThreshold {
		sqlSlowQueries.WithLabelValues(trace.Method).Inc()
	}
}
