package logger_v2

import (
	"context"
	"log/slog"
	"time"

	gorm "gorm.io/gorm/logger"
)

// LogLevel 日志级别
type LogLevel string

const (
	// DebugLevel debug级别
	DebugLevel LogLevel = "debug"
	// InfoLevel info级别
	InfoLevel LogLevel = "info"
	// WarnLevel warn级别
	WarnLevel LogLevel = "warn"
	// ErrorLevel error级别
	ErrorLevel LogLevel = "error"
)

// ILogger 日志接口
type ILogger interface {
	// Logger 返回底层的slog.Logger
	Logger() *slog.Logger
	LogMode(level gorm.LogLevel) gorm.Interface
	// Debug 调试日志
	Debug(ctx context.Context, msg string, data ...interface{})
	// Info 信息日志
	Info(ctx context.Context, msg string, data ...interface{})
	// Warn 警告日志
	Warn(ctx context.Context, msg string, data ...interface{})
	// Error 错误日志
	Error(ctx context.Context, msg string, data ...interface{})
	// Trace 追踪日志
	Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error)
	// DebugEnabled 是否启用debug级别
	DebugEnabled() bool
	// InfoEnabled 是否启用info级别
	InfoEnabled() bool
	// WarnEnabled 是否启用warn级别
	WarnEnabled() bool
	// ErrorEnabled 是否启用error级别
	ErrorEnabled() bool
	// With 添加属性
	With(attrs ...slog.Attr) *Logger
	// WithGroup 添加组
	WithGroup(name string) *Logger
}

// Config 日志配置
type Config struct {
	// Level 日志级别
	Level LogLevel `json:"level" yaml:"level" toml:"level"`
	// Format 日志格式 (json/text)
	Format string `json:"format" yaml:"format" toml:"format"`
	// FilePath 日志文件路径
	FilePath string `json:"file_path" yaml:"file_path" toml:"file_path"`
	// MaxSize 单个日志文件最大大小（MB）
	MaxSize int `json:"max_size" yaml:"max_size" toml:"max_size"`
	// MaxAge 日志文件保留天数
	MaxAge int `json:"max_age" yaml:"max_age" toml:"max_age"`
	// MaxBackups 日志文件最大保留数
	MaxBackups int `json:"max_backups" yaml:"max_backups" toml:"max_backups"`
	// Compress 是否压缩
	Compress bool `json:"compress" yaml:"compress" toml:"compress"`
	// DisableConsole 是否禁用控制台输出
	DisableConsole bool `json:"disable_console" yaml:"disable_console" toml:"disable_console"`
	// TimeLayout 时间格式
	TimeLayout string `json:"time_layout" yaml:"time_layout" toml:"time_layout"`
	// Fields 固定字段
	Fields map[string]string `json:"fields" yaml:"fields" toml:"fields"`
	// AddSource 是否添加源码位置信息
	AddSource bool `json:"add_source" yaml:"add_source" toml:"add_source"`
}

// ToSlogLevel converts LogLevel to slog.Level
func (l LogLevel) ToSlogLevel() slog.Level {
	switch l {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// NewFromConfig creates a new logger from config
func NewFromConfig(cfg Config) (*Logger, error) {
	var opts []Option

	// Set level
	switch cfg.Level.ToSlogLevel() {
	case slog.LevelDebug:
		opts = append(opts, WithDebugLevel())
	case slog.LevelInfo:
		opts = append(opts, WithInfoLevel())
	case slog.LevelWarn:
		opts = append(opts, WithWarnLevel())
	case slog.LevelError:
		opts = append(opts, WithErrorLevel())
	}

	// Set format
	if cfg.Format == "json" {
		opts = append(opts, WithFormatJSON())
	}

	// Set file output
	if cfg.FilePath != "" {
		if cfg.MaxSize > 0 || cfg.MaxAge > 0 || cfg.MaxBackups > 0 {
			opts = append(opts, WithFileRotationP(cfg.FilePath, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge))
		} else {
			opts = append(opts, WithFileP(cfg.FilePath))
		}
	}

	// Set console output
	if cfg.DisableConsole {
		opts = append(opts, WithDisableConsole())
	}

	// Set time layout
	if cfg.TimeLayout != "" {
		opts = append(opts, WithTimeLayout(cfg.TimeLayout))
	}

	// Set source
	if cfg.AddSource {
		opts = append(opts, WithSource())
	}

	// Set fields
	for k, v := range cfg.Fields {
		opts = append(opts, WithField(k, v))
	}

	return New(opts...)
}
