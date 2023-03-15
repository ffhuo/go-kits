package db

import (
	"strconv"

	"github.com/ffhuo/go-kits/custom_time"
	"github.com/prometheus/client_golang/prometheus"
	plugin "gorm.io/plugin/prometheus"
)

type Metric struct {
	Prefix string
	status *prometheus.HistogramVec
}

func (m *Metric) Metrics(p *plugin.Prometheus) []prometheus.Collector {
	if m.Prefix == "" {
		m.Prefix = "gorm_"
	}

	gramv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   m.Prefix + "sql",
		Subsystem:   "",
		Name:        m.Prefix + "sql_exec",
		ConstLabels: nil,
		Buckets:     nil,
	}, []string{"method", "time", "sql", "rows"})

	m.status = gramv
	prometheus.Register(gramv)
	return []prometheus.Collector{m.status}
}

func (m *Metric) Collect(trace *SQLTrace) {
	m.status.WithLabelValues(
		trace.Method,
		trace.Timestamp.Format(custom_time.TimeLayout),
		trace.SQL,
		strconv.Itoa(int(trace.Rows)),
	).Observe(trace.CostSeconds)
}
