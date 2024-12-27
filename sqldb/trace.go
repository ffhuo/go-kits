package sqldb

import (
	"errors"
	"time"

	"github.com/ffhuo/go-kits/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/utils"
)

const (
	callBackBeforeName = "trace:before"
	callBackAfterName  = "trace:after"
)

// TracePlugin SQL 追踪插件
type TracePlugin struct {
	metrics *SQLMetrics
	logger  logger.ILogger
}

// NewTrace 创建 SQL 追踪插件
func NewTrace(metrics *SQLMetrics, logger logger.ILogger) *TracePlugin {
	return &TracePlugin{
		metrics: metrics,
		logger:  logger,
	}
}

// Name 插件名称
func (op *TracePlugin) Name() string {
	return "tracePlugin"
}

// Initialize 初始化插件
func (op *TracePlugin) Initialize(db *gorm.DB) error {
	// 注册回调
	if err := db.Callback().Create().Before("gorm:create").Register(callBackBeforeName, op.before); err != nil {
		return err
	}
	if err := db.Callback().Create().After("gorm:create").Register(callBackAfterName, op.after); err != nil {
		return err
	}

	if err := db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, op.before); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gorm:query").Register(callBackAfterName, op.after); err != nil {
		return err
	}

	if err := db.Callback().Delete().Before("gorm:delete").Register(callBackBeforeName, op.before); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:delete").Register(callBackAfterName, op.after); err != nil {
		return err
	}

	if err := db.Callback().Update().Before("gorm:update").Register(callBackBeforeName, op.before); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register(callBackAfterName, op.after); err != nil {
		return err
	}

	if err := db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, op.before); err != nil {
		return err
	}
	if err := db.Callback().Row().After("gorm:row").Register(callBackAfterName, op.after); err != nil {
		return err
	}

	if err := db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, op.before); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, op.after); err != nil {
		return err
	}

	return nil
}

func (op *TracePlugin) before(db *gorm.DB) {
	db.InstanceSet("gorm:trace_start_time", time.Now())
}

func (op *TracePlugin) after(db *gorm.DB) {
	_ts, ok := db.InstanceGet("gorm:trace_start_time")
	if !ok {
		return
	}

	ts, ok := _ts.(time.Time)
	if !ok {
		return
	}

	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	method := utils.FileWithLineNum()

	// 收集指标
	trace := &SQLTrace{
		Method:      method,
		SQL:         sql,
		Stack:       utils.FileWithLineNum(),
		Rows:        db.Statement.RowsAffected,
		CostSeconds: time.Since(ts).Seconds(),
		Timestamp:   ts,
	}
	op.metrics.Collect(trace)

	// 记录日志
	fields := []interface{}{
		"sql", sql,
		"rows", trace.Rows,
		"duration", time.Duration(trace.CostSeconds * float64(time.Second)),
	}

	if db.Error != nil && !errors.Is(db.Error, gorm.ErrRecordNotFound) {
		fields = append(fields, "error", db.Error)
		op.logger.Error(db.Statement.Context, "SQL执行出错", fields...)
	} else {
		op.logger.Debug(db.Statement.Context, "SQL执行成功", fields...)
	}
}

// SQLTrace SQL 执行跟踪信息
type SQLTrace struct {
	Method      string
	SQL         string
	Stack       string
	Rows        int64
	CostSeconds float64
	Timestamp   time.Time
}
