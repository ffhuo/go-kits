package db

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/utils"
)

const (
	callBackBeforeName = "core:before"
	callBackAfterName  = "core:after"
	startTime          = "_start_time"
)

type TracePlugin struct {
	m *Metric
}

func NewTrace(m *Metric) *TracePlugin {
	return &TracePlugin{
		m: m,
	}
}

func (op *TracePlugin) Name() string {
	return "tracePlugin"
}

func (op *TracePlugin) Initialize(db *gorm.DB) (err error) {
	// 开始前
	_ = db.Callback().Create().Before("gorm:before_create").Register(callBackBeforeName, op.before)
	_ = db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, op.before)
	_ = db.Callback().Delete().Before("gorm:before_delete").Register(callBackBeforeName, op.before)
	_ = db.Callback().Update().Before("gorm:setup_reflect_value").Register(callBackBeforeName, op.before)
	_ = db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, op.before)
	_ = db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, op.before)

	// 结束后
	_ = db.Callback().Create().After("gorm:after_create").Register(callBackAfterName, op.createAfter)
	_ = db.Callback().Query().After("gorm:after_query").Register(callBackAfterName, op.queryAfter)
	_ = db.Callback().Delete().After("gorm:after_delete").Register(callBackAfterName, op.deleteAfter)
	_ = db.Callback().Update().After("gorm:after_update").Register(callBackAfterName, op.updateAfter)
	_ = db.Callback().Row().After("gorm:row").Register(callBackAfterName, op.rowAfter)
	_ = db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, op.rawAfter)
	return
}

var _ gorm.Plugin = &TracePlugin{}

func (op *TracePlugin) before(db *gorm.DB) {
	db.InstanceSet(startTime, time.Now())
}

func (op *TracePlugin) createAfter(db *gorm.DB) {
	op.after("CREATE", db)
}
func (op *TracePlugin) queryAfter(db *gorm.DB) {
	op.after("QUERY", db)
}
func (op *TracePlugin) deleteAfter(db *gorm.DB) {
	op.after("DELETE", db)
}
func (op *TracePlugin) updateAfter(db *gorm.DB) {
	op.after("UPDATE", db)
}
func (op *TracePlugin) rowAfter(db *gorm.DB) {
	op.after("ROW", db)
}
func (op *TracePlugin) rawAfter(db *gorm.DB) {
	op.after("RAW", db)
}
func (op *TracePlugin) after(method string, db *gorm.DB) {
	if op.m == nil {
		return
	}
	_ts, isExist := db.InstanceGet(startTime)
	if !isExist {
		return
	}

	ts, ok := _ts.(time.Time)
	if !ok {
		return
	}

	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)

	sqlInfo := new(SQLTrace)
	sqlInfo.Method = method
	sqlInfo.Timestamp = time.Now()
	sqlInfo.SQL = sql
	sqlInfo.Stack = utils.FileWithLineNum()
	sqlInfo.Rows = db.Statement.RowsAffected
	sqlInfo.CostSeconds = time.Since(ts).Seconds()
	op.m.Collect(sqlInfo)
}

type SQLTrace struct {
	Method      string
	SQL         string
	Stack       string
	Rows        int64
	CostSeconds float64
	Timestamp   time.Time
}
