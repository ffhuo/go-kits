package db

import (
	"context"
	"time"

	"github.com/ffhuo/go-conf/pkg/paginator"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Option options
type Option func(*option)

type option struct {
	driver          string
	rAddr           string
	wAddr           string
	debug           bool
	maxOpenConn     int
	maxIdleConn     int
	connMaxLifeTime time.Duration
}

func Debug() Option {
	return func(opt *option) {
		opt.debug = true
	}
}

func SetDriverMysql() Option {
	return func(opt *option) {
		opt.driver = "mysql"
	}
}

func SetDriverSqlite() Option {
	return func(opt *option) {
		opt.driver = "sqlite"
	}
}

func SetReadAddr(addr string) Option {
	return func(opt *option) {
		opt.rAddr = addr
	}
}

func SetWriteAddr(addr string) Option {
	return func(opt *option) {
		opt.wAddr = addr
	}
}

func SetMaxOpenConns(maxOpenConn int) Option {
	return func(opt *option) {
		// sqlDB, err := cli.DB()
		// if err != nil {
		// 	return
		// }

		// // 设置连接池 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
		// sqlDB.SetMaxOpenConns(maxOpenConn)
		opt.maxOpenConn = maxOpenConn
	}
}

func SetMaxIdleConns(maxIdleConn int) Option {
	return func(opt *option) {
		opt.maxIdleConn = maxIdleConn
		// sqlDB, err := cli.DB()
		// if err != nil {
		// 	return
		// }

		// // 设置最大连接数 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
		// sqlDB.SetMaxIdleConns(maxIdleConn)
	}
}

func SetConnMaxLifetime(connMaxLifeTime time.Duration) Option {
	return func(opt *option) {
		opt.connMaxLifeTime = connMaxLifeTime
		// sqlDB, err := cli.DB()
		// if err != nil {
		// 	return
		// }

		// // 设置最大连接超时
		// sqlDB.SetConnMaxLifetime(time.Minute * connMaxLifeTime)
	}
}

type Client struct {
	debug bool
	dbR   *gorm.DB
	dbW   *gorm.DB
}

func New(opts ...Option) (*Client, error) {
	var (
		err error
		cli Client = Client{}
	)

	var opt option
	for _, o := range opts {
		o(&opt)
	}

	if opt.rAddr == "" && opt.wAddr == "" {
		return nil, errors.New("read and write addr is empty")
	}

	if opt.rAddr != "" {
		if cli.dbR, err = DBConnect(opt.rAddr, &opt); err != nil {
			return nil, errors.Wrap(err, "connect read db error")
		}
	}

	if opt.wAddr != "" {
		if cli.dbW, err = DBConnect(opt.wAddr, &opt); err != nil {
			return nil, errors.Wrap(err, "connect write db error")
		}
	}

	return &cli, nil
}

func (cli *Client) GetDbReader(ctx context.Context) *gorm.DB {
	if cli.dbR == nil {
		return cli.GetDbWriter(ctx)
	}

	if cli.debug {
		return cli.dbR.WithContext(ctx).Debug()
	}
	return cli.dbR.WithContext(ctx)
}

func (cli *Client) GetDbWriter(ctx context.Context) *gorm.DB {
	if cli.dbW == nil {
		return cli.GetDbReader(ctx)
	}
	if cli.debug {
		return cli.dbW.WithContext(ctx).Debug()
	}
	return cli.dbW.WithContext(ctx)
}

func (cli *Client) Save(ctx context.Context, value interface{}) error {
	return cli.GetDbWriter(ctx).Save(value).Error
}

func (cli *Client) Create(ctx context.Context, value interface{}) error {
	return cli.GetDbWriter(ctx).Create(value).Error
}

type Args func(*gorm.DB) *gorm.DB

func Where(query interface{}, args ...interface{}) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Where(query, args...)
	}
}

func Order(query interface{}) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Order(query)
	}
}

func Unscoped() Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Unscoped()
	}
}

func (cli *Client) Delete(ctx context.Context, model interface{}, args ...Args) error {
	dbCli := cli.GetDbWriter(ctx)
	dbCli = dbCli.Model(model)
	for _, arg := range args {
		dbCli = arg(dbCli)
	}
	return dbCli.Delete(model).Error
}

func (cli *Client) Find(ctx context.Context, data interface{}, pagination *paginator.Paginator, args ...Args) (int64, error) {
	var total int64
	dbCli := cli.GetDbReader(ctx)
	dbCli = dbCli.Model(data)
	for _, arg := range args {
		dbCli = arg(dbCli)
	}
	if pagination != nil {
		dbCli.Count(&total)
		dbCli = dbCli.Limit(pagination.GetLimit()).Offset(pagination.GetOffset())
	}
	dbCli = dbCli.Find(data)
	return total, dbCli.Error
}

func (cli *Client) FindOne(ctx context.Context, data interface{}, args ...Args) error {
	dbCli := cli.GetDbReader(ctx)
	dbCli = dbCli.Model(data)
	for _, arg := range args {
		dbCli = arg(dbCli)
	}
	dbCli = dbCli.First(data)
	return dbCli.Error
}

func (cli *Client) Count(ctx context.Context, model interface{}, args ...Args) (int64, error) {
	var total int64
	dbCli := cli.GetDbReader(ctx)
	dbCli = dbCli.Model(model)
	for _, arg := range args {
		dbCli = arg(dbCli)
	}
	dbCli = dbCli.Count(&total)
	return total, dbCli.Error
}

func (cli *Client) Update(ctx context.Context, model interface{}, metric map[string]interface{}, args ...Args) error {
	dbCli := cli.GetDbWriter(ctx)
	dbCli = dbCli.Model(model)
	for _, arg := range args {
		dbCli = arg(dbCli)
	}

	return dbCli.Updates(metric).Error
}

func DBConnect(addr string, opt *option) (*gorm.DB, error) {
	switch opt.driver {
	case "mysql":
		return MysqlConnect(addr, opt)
	default:
		return SqliteConnect(addr, opt)
	}
}

// dbConnect: 初始化数据库
func MysqlConnect(addr string, opt *option) (*gorm.DB, error) {
	gormConfig := &gorm.Config{}

	cli, err := gorm.Open(mysql.New(mysql.Config{
		DSN: addr,
	}), gormConfig)
	if err != nil {
		return nil, errors.Wrap(err, "初始化DB时出错")
	}

	sqlDB, err := cli.DB()
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect")
	}

	if opt.connMaxLifeTime != 0 {
		sqlDB.SetConnMaxLifetime(time.Minute * opt.connMaxLifeTime)
	}

	if opt.maxIdleConn != 0 {
		sqlDB.SetMaxIdleConns(opt.maxIdleConn)
	}

	if opt.maxOpenConn != 0 {
		sqlDB.SetMaxOpenConns(opt.maxOpenConn)
	}

	return cli, nil
}

func SqliteConnect(addr string, opt *option) (*gorm.DB, error) {
	gormConfig := &gorm.Config{}

	cli, err := gorm.Open(sqlite.Open(addr), gormConfig)
	if err != nil {
		return nil, errors.Wrap(err, "初始化DB时出错")
	}

	sqlDB, err := cli.DB()
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect")
	}

	if opt.connMaxLifeTime != 0 {
		sqlDB.SetConnMaxLifetime(time.Minute * opt.connMaxLifeTime)
	}

	if opt.maxIdleConn != 0 {
		sqlDB.SetMaxIdleConns(opt.maxIdleConn)
	}

	if opt.maxOpenConn != 0 {
		sqlDB.SetMaxOpenConns(opt.maxOpenConn)
	}

	return cli, nil
}
