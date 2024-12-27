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
	// DefaultTimeLayout the default time layout
	DefaultTimeLayout = time.RFC3339
	// DefaultMaxSize the default max size of log file (MB)
	DefaultMaxSize = 100
	// DefaultMaxAge the default max age of log file (days)
	DefaultMaxAge = 7
	// DefaultMaxBackups the default max backups of log file
	DefaultMaxBackups = 10
	// SlowThreshold the default slow threshold
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

// WithFormatJSON set json format
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
func WithFileRotationP(file string, maxSize, maxBackups, maxAge int) Option {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0766); err != nil {
		panic(err)
	}

	if maxSize <= 0 {
		maxSize = DefaultMaxSize
	}
	if maxBackups <= 0 {
		maxBackups = DefaultMaxBackups
	}
	if maxAge <= 0 {
		maxAge = DefaultMaxAge
	}

	return func(opt *option) {
		opt.file = &lumberjack.Logger{
			Filename:   file,
			MaxSize:    maxSize,    // MB
			MaxBackups: maxBackups, // files
			MaxAge:     maxAge,     // days
			LocalTime:  true,
			Compress:   true,
		}
	}
}

// WithTimeLayout custom time format
func WithTimeLayout(timeLayout string) Option {
	return func(opt *option) {
		opt.timeLayout = timeLayout
	}
}

// WithDisableConsole disable console output
func WithDisableConsole() Option {
	return func(opt *option) {
		opt.disableConsole = true
	}
}

type Logger struct {
	log *zap.Logger
}

// New create a new logger
func New(opts ...Option) (*Logger, error) {
	opt := &option{
		level:  DefaultLevel,
		fields: make(map[string]string),
	}
	for _, f := range opts {
		f(opt)
	}

	timeLayout := DefaultTimeLayout
	if opt.timeLayout != "" {
		timeLayout = opt.timeLayout
	}

	// 自定义 Duration 编码器
	durationEncoder := func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
		switch {
		case d < time.Microsecond:
			enc.AppendString(fmt.Sprintf("%dns", d.Nanoseconds()))
		case d < time.Millisecond:
			enc.AppendString(fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/float64(time.Microsecond)))
		case d < time.Second:
			enc.AppendString(fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/float64(time.Millisecond)))
		default:
			enc.AppendString(fmt.Sprintf("%.2fs", float64(d.Nanoseconds())/float64(time.Second)))
		}
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(timeLayout),
		EncodeDuration: durationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if opt.formatJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// lowPriority used by info\debug\warn
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl < zapcore.ErrorLevel
	})

	// highPriority used by error\panic\fatal
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl >= zapcore.ErrorLevel
	})

	var cores []zapcore.Core

	if !opt.disableConsole {
		stdout := zapcore.Lock(os.Stdout)
		stderr := zapcore.Lock(os.Stderr)

		cores = append(cores,
			zapcore.NewCore(encoder, stdout, lowPriority),
			zapcore.NewCore(encoder, stderr, highPriority),
		)
	}

	if opt.file != nil {
		cores = append(cores,
			zapcore.NewCore(encoder, zapcore.AddSync(opt.file),
				zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					return lvl >= opt.level
				}),
			),
		)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	if len(opt.fields) > 0 {
		fields := make([]zap.Field, 0, len(opt.fields))
		for k, v := range opt.fields {
			fields = append(fields, zap.String(k, v))
		}
		logger = logger.With(fields...)
	}

	return &Logger{log: logger}, nil
}

func (l *Logger) Logger() *zap.Logger {
	return l.log
}

func (l *Logger) LogMode(level gorm.LogLevel) gorm.Interface {
	return l
}

func (l *Logger) Debug(ctx context.Context, msg string, data ...interface{}) {
	if fields := Fields(ctx); len(fields) > 0 {
		l.log.With(fields...).Debug(fmt.Sprintf(msg, data...))
	} else {
		l.log.Debug(fmt.Sprintf(msg, data...))
	}
}

func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if fields := Fields(ctx); len(fields) > 0 {
		l.log.With(fields...).Info(fmt.Sprintf(msg, data...))
	} else {
		l.log.Info(fmt.Sprintf(msg, data...))
	}
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if fields := Fields(ctx); len(fields) > 0 {
		l.log.With(fields...).Warn(fmt.Sprintf(msg, data...))
	} else {
		l.log.Warn(fmt.Sprintf(msg, data...))
	}
}

func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if fields := Fields(ctx); len(fields) > 0 {
		l.log.With(fields...).Error(fmt.Sprintf(msg, data...))
	} else {
		l.log.Error(fmt.Sprintf(msg, data...))
	}
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	baseFields := Fields(ctx)
	fields := make([]zap.Field, 0, len(baseFields)+5)
	fields = append(fields, baseFields...)

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		fields = append(fields,
			zap.String("line", utils.FileWithLineNum()),
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
		l.log.Error("trace", fields...)
	case elapsed > SlowThreshold:
		fields = append(fields,
			zap.String("line", utils.FileWithLineNum()),
			zap.String("slow", fmt.Sprintf(">= %v", SlowThreshold)),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
		l.log.Warn("trace", fields...)
	default:
		fields = append(fields,
			zap.String("line", utils.FileWithLineNum()),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
		l.log.Debug("trace", fields...)
	}
}
