package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	gorm "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const (
	// DefaultLevel the default log level
	DefaultLevel = zapcore.InfoLevel

	// DefaultTimeLayout the default time layout;
	DefaultTimeLayout = time.RFC3339

	SlowThreshold = time.Millisecond * 200
)

// Option custom setup config
type Option func(*option)

type option struct {
	level          zapcore.Level
	fields         map[string]string
	file           io.Writer
	timeLayout     string
	disableConsole bool
	formatJSON     bool
}

func WithFormatJSON() Option {
	return func(opt *option) {
		opt.formatJSON = true
	}
}

// WithDebugLevel only greater than 'level' will output
func WithDebugLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.DebugLevel
	}
}

// WithInfoLevel only greater than 'level' will output
func WithInfoLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.InfoLevel
	}
}

// WithWarnLevel only greater than 'level' will output
func WithWarnLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.WarnLevel
	}
}

// WithErrorLevel only greater than 'level' will output
func WithErrorLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.ErrorLevel
	}
}

// WithField add some field(s) to log
func WithField(key, value string) Option {
	return func(opt *option) {
		opt.fields[key] = value
	}
}

// WithFileP write log to some file
func WithFileP(file string) Option {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0766); err != nil {
		panic(err)
	}

	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0766)
	if err != nil {
		panic(err)
	}

	return func(opt *option) {
		opt.file = zapcore.Lock(f)
	}
}

// WithFileRotationP write log to some file with rotation
func WithFileRotationP(file string) Option {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0766); err != nil {
		panic(err)
	}

	return func(opt *option) {
		opt.file = &lumberjack.Logger{ // concurrent-safed
			Filename:   file, // 文件路径
			MaxSize:    128,  // 单个文件最大尺寸，默认单位 M
			MaxBackups: 300,  // 最多保留 300 个备份
			MaxAge:     30,   // 最大时间，默认单位 day
			LocalTime:  true, // 使用本地时间
			Compress:   true, // 是否压缩 disabled by default
		}
	}
}

// WithTimeLayout custom time format
func WithTimeLayout(timeLayout string) Option {
	return func(opt *option) {
		opt.timeLayout = timeLayout
	}
}

// WithEnableConsole write log to os.Stdout or os.Stderr
func WithDisableConsole() Option {
	return func(opt *option) {
		opt.disableConsole = true
	}
}

type Logger struct {
	log *zap.Logger
}

func New(opts ...Option) (*Logger, error) {
	log, err := NewLogger(opts...)
	if err != nil {
		return nil, err
	}
	return &Logger{log: log}, nil
}

func (l *Logger) Logger() *zap.Logger {
	return l.log
}

func (l *Logger) Debug(ctx context.Context, msg string, data ...interface{}) {
	log := l.log.Sugar()
	fileds := Fields(ctx)
	if len(fileds) != 0 {
		log = log.With(fileds)
	}
	log.Debugf(msg, data...)
}

func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	log := l.log.Sugar()
	fileds := Fields(ctx)
	if len(fileds) != 0 {
		log = log.With(fileds)
	}
	log.Infof(msg, data...)
}

func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	log := l.log.Sugar()
	fileds := Fields(ctx)
	if len(fileds) != 0 {
		log = log.With(fileds)
	}
	log.Errorf(msg, data...)
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	log := l.log.Sugar()
	fileds := Fields(ctx)
	if len(fileds) != 0 {
		log = log.With(fileds)
	}
	log.Warnf(msg, data...)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	log := l.log.Sugar()
	fileds := Fields(ctx)
	if len(fileds) != 0 {
		log = log.With(fileds)
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && gorm.LogLevel(DefaultLevel) >= gorm.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		if rows == -1 {
			log.Errorf("%s %s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			log.Errorf("%s %s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > SlowThreshold && SlowThreshold != 0 && gorm.LogLevel(DefaultLevel) >= gorm.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", SlowThreshold)
		if rows == -1 {
			log.Warnf("%s %s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			log.Warnf("%s %s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case gorm.LogLevel(DefaultLevel) == gorm.Info:
		sql, rows := fc()
		if rows == -1 {
			log.Infof("%s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			log.Infof("%s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

// NewJSONLogger return a json-encoder zap logger,
func NewLogger(opts ...Option) (*zap.Logger, error) {
	opt := &option{level: DefaultLevel, fields: make(map[string]string)}
	for _, f := range opts {
		f(opt)
	}

	timeLayout := DefaultTimeLayout
	if opt.timeLayout != "" {
		timeLayout = opt.timeLayout
	}

	// similar to zap.NewProductionEncoderConfig()
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger", // used by logger.Named(key); optional; useless
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace", // use by zap.AddStacktrace; optional; useless
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(timeLayout))
		},
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
	}

	var encoder zapcore.Encoder
	if opt.formatJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// lowPriority usd by info\debug\warn
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl < zapcore.ErrorLevel
	})

	// highPriority usd by error\panic\fatal
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl >= zapcore.ErrorLevel
	})

	stdout := zapcore.Lock(os.Stdout) // lock for concurrent safe
	stderr := zapcore.Lock(os.Stderr) // lock for concurrent safe

	core := zapcore.NewTee()

	if !opt.disableConsole {
		core = zapcore.NewTee(
			zapcore.NewCore(encoder,
				zapcore.NewMultiWriteSyncer(stdout),
				lowPriority,
			),
			zapcore.NewCore(encoder,
				zapcore.NewMultiWriteSyncer(stderr),
				highPriority,
			),
		)
	}

	if opt.file != nil {
		core = zapcore.NewTee(core,
			zapcore.NewCore(encoder,
				zapcore.AddSync(opt.file),
				zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					return lvl >= opt.level
				}),
			),
		)
	}

	logger := zap.New(core,
		zap.AddCaller(),
		zap.ErrorOutput(stderr),
	)

	for key, value := range opt.fields {
		logger = logger.WithOptions(zap.Fields(zapcore.Field{Key: key, Type: zapcore.StringType, String: value}))
	}

	return logger, nil
}
