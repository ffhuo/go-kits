package sqldb

import (
	"context"
	"fmt"
	"time"

	"github.com/ffhuo/go-kits/logger"
	"github.com/ffhuo/go-kits/paginator"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Option 配置选项
type Option func(*options)

type options struct {
	// 数据库类型
	driver string

	// 连接地址
	master string   // 主库地址
	slaves []string // 从库地址列表

	// 连接池配置
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration

	// 高级特性
	debug         bool          // 调试模式
	slowThreshold time.Duration // 慢查询阈值
	logger        *logger.Logger
	plugins       []gorm.Plugin
}

// WithMySQL 使用 MySQL 数据库
func WithMySQL() Option {
	return func(opt *options) {
		opt.driver = "mysql"
	}
}

// WithSQLite 使用 SQLite 数据库
func WithSQLite() Option {
	return func(opt *options) {
		opt.driver = "sqlite"
	}
}

// WithClickHouse 使用 ClickHouse 数据库
func WithClickHouse() Option {
	return func(opt *options) {
		opt.driver = "clickhouse"
	}
}

// WithMaster 设置主库地址
func WithMaster(dsn string) Option {
	return func(opt *options) {
		opt.master = dsn
	}
}

// WithSlaves 设置从库地址
func WithSlaves(dsn ...string) Option {
	return func(opt *options) {
		opt.slaves = dsn
	}
}

// WithMaxOpenConns 设置最大连接数
func WithMaxOpenConns(n int) Option {
	return func(opt *options) {
		opt.maxOpenConns = n
	}
}

// WithMaxIdleConns 设置最大空闲连接数
func WithMaxIdleConns(n int) Option {
	return func(opt *options) {
		opt.maxIdleConns = n
	}
}

// WithConnMaxLifetime 设置连接最大生命周期
func WithConnMaxLifetime(d time.Duration) Option {
	return func(opt *options) {
		opt.connMaxLifetime = d
	}
}

// WithConnMaxIdleTime 设置空闲连接最大生命周期
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(opt *options) {
		opt.connMaxIdleTime = d
	}
}

// WithDebug 开启调试模式
func WithDebug() Option {
	return func(opt *options) {
		opt.debug = true
	}
}

// WithSlowThreshold 设置慢查询阈值
func WithSlowThreshold(d time.Duration) Option {
	return func(opt *options) {
		opt.slowThreshold = d
	}
}

// WithLogger 设置日志记录器
func WithLogger(log *logger.Logger) Option {
	return func(opt *options) {
		opt.logger = log
	}
}

// WithPlugin 添加插件
func WithPlugin(plugins ...gorm.Plugin) Option {
	return func(opt *options) {
		opt.plugins = append(opt.plugins, plugins...)
	}
}

// DB 数据库客户端
type DB struct {
	master *gorm.DB   // 主库连接
	slaves []*gorm.DB // 从库连接
	opts   *options   // 配置选项
}

// New 创建数据库客户端
func New(opts ...Option) (*DB, error) {
	options := &options{
		maxOpenConns:    100,
		maxIdleConns:    10,
		connMaxLifetime: time.Hour,
		connMaxIdleTime: time.Hour,
		slowThreshold:   time.Second,
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.master == "" {
		return nil, fmt.Errorf("master DSN is required")
	}

	db := &DB{opts: options}

	// 连接主库
	master, err := connect(options.driver, options.master, options)
	if err != nil {
		return nil, fmt.Errorf("connect master failed: %v", err)
	}
	db.master = master

	// 连接从库
	for _, slave := range options.slaves {
		s, err := connect(options.driver, slave, options)
		if err != nil {
			return nil, fmt.Errorf("connect slave failed: %v", err)
		}
		db.slaves = append(db.slaves, s)
	}

	return db, nil
}

// Write 获取写库连接
func (db *DB) Write(ctx context.Context) *gorm.DB {
	if db.opts.debug {
		return db.master.WithContext(ctx).Debug()
	}
	return db.master.WithContext(ctx)
}

// Read 获取读库连接
func (db *DB) Read(ctx context.Context) *gorm.DB {
	if len(db.slaves) == 0 {
		return db.Write(ctx)
	}

	// TODO: 实现负载均衡
	slave := db.slaves[0]
	if db.opts.debug {
		return slave.WithContext(ctx).Debug()
	}
	return slave.WithContext(ctx)
}

// connect 连接数据库
func connect(driver, dsn string, opt *options) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch driver {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	case "clickhouse":
		dialector = clickhouse.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	config := &gorm.Config{}
	if opt.logger != nil {
		config.Logger = opt.logger
	}

	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池
	sqlDB.SetMaxOpenConns(opt.maxOpenConns)
	sqlDB.SetMaxIdleConns(opt.maxIdleConns)
	sqlDB.SetConnMaxLifetime(opt.connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(opt.connMaxIdleTime)

	// 注册插件
	for _, plugin := range opt.plugins {
		if err := db.Use(plugin); err != nil {
			return nil, fmt.Errorf("register plugin failed: %v", err)
		}
	}

	return db, nil
}

// Transaction 事务
func (db *DB) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return db.Write(ctx).Transaction(fn)
}

// Create 创建记录
func (db *DB) Create(ctx context.Context, value interface{}) error {
	return db.Write(ctx).Create(value).Error
}

// Save 保存记录
func (db *DB) Save(ctx context.Context, value interface{}) error {
	return db.Write(ctx).Save(value).Error
}

// Delete 删除记录
func (db *DB) Delete(ctx context.Context, value interface{}, args ...Args) error {
	tx := db.Write(ctx)
	for _, arg := range args {
		tx = arg(tx)
	}
	return tx.Delete(value).Error
}

// Update 更新记录
func (db *DB) Update(ctx context.Context, model interface{}, values interface{}, args ...Args) error {
	tx := db.Write(ctx)
	for _, arg := range args {
		tx = arg(tx)
	}
	return tx.Model(model).Updates(values).Error
}

// First 查询单条记录
func (db *DB) First(ctx context.Context, dest interface{}, args ...Args) error {
	tx := db.Read(ctx)
	for _, arg := range args {
		tx = arg(tx)
	}
	return tx.First(dest).Error
}

// Find 查询多条记录
func (db *DB) Find(ctx context.Context, data interface{}, pagination *paginator.Paginator, args ...Args) (int64, error) {
	var total int64
	dbCli := db.Read(ctx)
	dbCli = dbCli.Model(data)
	for _, arg := range args {
		dbCli = arg(dbCli)
	}
	if pagination != nil {
		dbCli.Count(&total)
		dbCli = dbCli.Limit(pagination.Limit()).Offset(pagination.Offset())
	}
	dbCli = dbCli.Find(data)
	return total, dbCli.Error
}

// Count 统计记录数
func (db *DB) Count(ctx context.Context, model interface{}, args ...Args) (int64, error) {
	var count int64
	tx := db.Read(ctx)
	for _, arg := range args {
		tx = arg(tx)
	}
	err := tx.Model(model).Count(&count).Error
	return count, err
}

// Close 关闭连接
func (db *DB) Close() error {
	if db.master != nil {
		sqlDB, err := db.master.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.Close(); err != nil {
			return err
		}
	}

	for _, slave := range db.slaves {
		sqlDB, err := slave.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.Close(); err != nil {
			return err
		}
	}

	return nil
}
