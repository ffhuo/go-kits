package logger_v2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
	gorm "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const (
	// DefaultLevel the default log level
	DefaultLevel = slog.LevelInfo
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
	level          slog.Level
	fields         map[string]string
	file           io.Writer
	timeLayout     string
	disableConsole bool
	formatJSON     bool
	addSource      bool
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
		opt.level = slog.LevelDebug
	}
}

// WithInfoLevel only greater than 'level' will output
func WithInfoLevel() Option {
	return func(opt *option) {
		opt.level = slog.LevelInfo
	}
}

// WithWarnLevel only greater than 'level' will output
func WithWarnLevel() Option {
	return func(opt *option) {
		opt.level = slog.LevelWarn
	}
}

// WithErrorLevel only greater than 'level' will output
func WithErrorLevel() Option {
	return func(opt *option) {
		opt.level = slog.LevelError
	}
}

// WithSource add source information (file:line)
func WithSource() Option {
	return func(opt *option) {
		opt.addSource = true
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
		opt.file = f
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

// Logger wraps slog.Logger
type Logger struct {
	logger *slog.Logger
	opts   *option
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

	var writers []io.Writer

	// Add console output if not disabled
	if !opt.disableConsole {
		writers = append(writers, os.Stdout)
	}

	// Add file output if specified
	if opt.file != nil {
		writers = append(writers, opt.file)
	}

	// If no writers specified, use stdout
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	// Create multi-writer
	var writer io.Writer
	if len(writers) == 1 {
		writer = writers[0]
	} else {
		writer = io.MultiWriter(writers...)
	}

	// Create handler options
	handlerOpts := &slog.HandlerOptions{
		Level:     opt.level,
		AddSource: opt.addSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Custom time format
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.String(slog.TimeKey, t.Format(timeLayout))
				}
			}
			return a
		},
	}

	// Create handler
	var handler slog.Handler
	if opt.formatJSON {
		handler = slog.NewJSONHandler(writer, handlerOpts)
	} else {
		handler = slog.NewTextHandler(writer, handlerOpts)
	}

	// Create logger
	logger := slog.New(handler)

	// Add default fields
	if len(opt.fields) > 0 {
		args := make([]any, 0, len(opt.fields)*2)
		for k, v := range opt.fields {
			args = append(args, k, v)
		}
		logger = logger.With(args...)
	}

	return &Logger{
		logger: logger,
		opts:   opt,
	}, nil
}

// Logger returns the underlying slog.Logger
func (l *Logger) Logger() *slog.Logger {
	return l.logger
}

// LogMode implements gorm.Interface for GORM compatibility
func (l *Logger) LogMode(level gorm.LogLevel) gorm.Interface {
	return l
}

// Debug logs at debug level
func (l *Logger) Debug(ctx context.Context, msg string, data ...interface{}) {
	if fields := Fields(ctx); len(fields) > 0 {
		l.logger.DebugContext(ctx, fmt.Sprintf(msg, data...), fields...)
	} else {
		l.logger.DebugContext(ctx, fmt.Sprintf(msg, data...))
	}
}

// Info logs at info level
func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if fields := Fields(ctx); len(fields) > 0 {
		l.logger.InfoContext(ctx, fmt.Sprintf(msg, data...), fields...)
	} else {
		l.logger.InfoContext(ctx, fmt.Sprintf(msg, data...))
	}
}

// Warn logs at warn level
func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if fields := Fields(ctx); len(fields) > 0 {
		l.logger.WarnContext(ctx, fmt.Sprintf(msg, data...), fields...)
	} else {
		l.logger.WarnContext(ctx, fmt.Sprintf(msg, data...))
	}
}

// Error logs at error level
func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if fields := Fields(ctx); len(fields) > 0 {
		l.logger.ErrorContext(ctx, fmt.Sprintf(msg, data...), fields...)
	} else {
		l.logger.ErrorContext(ctx, fmt.Sprintf(msg, data...))
	}
}

// Trace implements gorm.Interface for GORM SQL tracing
func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	baseFields := Fields(ctx)
	fields := make([]any, 0, len(baseFields)*2+10) // each field becomes key-value pair

	// Add base fields
	for _, field := range baseFields {
		fields = append(fields, field)
	}

	// Add trace-specific fields
	fields = append(fields,
		slog.String("line", utils.FileWithLineNum()),
		slog.Duration("elapsed", elapsed),
		slog.Int64("rows", rows),
		slog.String("sql", sql),
	)

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		fields = append(fields, slog.Any("error", err))
		l.logger.ErrorContext(ctx, "trace", fields...)
	case elapsed > SlowThreshold:
		fields = append(fields, slog.String("slow", fmt.Sprintf(">= %v", SlowThreshold)))
		l.logger.WarnContext(ctx, "trace", fields...)
	default:
		l.logger.DebugContext(ctx, "trace", fields...)
	}
}

// DebugEnabled returns true if debug level is enabled
func (l *Logger) DebugEnabled() bool {
	return l.logger.Enabled(context.Background(), slog.LevelDebug)
}

// InfoEnabled returns true if info level is enabled
func (l *Logger) InfoEnabled() bool {
	return l.logger.Enabled(context.Background(), slog.LevelInfo)
}

// WarnEnabled returns true if warn level is enabled
func (l *Logger) WarnEnabled() bool {
	return l.logger.Enabled(context.Background(), slog.LevelWarn)
}

// ErrorEnabled returns true if error level is enabled
func (l *Logger) ErrorEnabled() bool {
	return l.logger.Enabled(context.Background(), slog.LevelError)
}

// With returns a new logger with the given attributes
func (l *Logger) With(attrs ...slog.Attr) *Logger {
	args := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value.Any())
	}
	return &Logger{
		logger: l.logger.With(args...),
		opts:   l.opts,
	}
}

// WithGroup returns a new logger with the given group name
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		logger: l.logger.WithGroup(name),
		opts:   l.opts,
	}
}
