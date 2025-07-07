package logger_v2

import (
	"context"
	"log/slog"
)

var MetaKey = "loggerMeta"
var _ Meta = (*meta)(nil)

// Meta key-value interface
type Meta interface {
	Key() string
	Value() interface{}
	meta()
}

type meta struct {
	key   string
	value interface{}
}

func (m *meta) Key() string {
	return m.key
}

func (m *meta) Value() interface{} {
	return m.value
}

func (m *meta) meta() {}

// NewMeta create meta
func NewMeta(key string, value interface{}) Meta {
	return &meta{key: key, value: value}
}

// WrapMeta wrap meta to slog attributes
func WrapMeta(err error, metas ...Meta) []any {
	capacity := len(metas) * 2 // each meta becomes key-value pair
	if err != nil {
		capacity += 2 // error key-value pair
	}

	fields := make([]any, 0, capacity)

	if err != nil {
		fields = append(fields, slog.Any("error", err))
	}

	// Add meta fields directly
	for _, meta := range metas {
		fields = append(fields, meta.Key(), meta.Value())
	}

	return fields
}

// Fields extracts fields from context and returns them as slog attributes
func Fields(ctx context.Context) []any {
	// Try to get gin context if available (optional dependency)
	if ginCtx, ok := ctx.(interface {
		Get(key string) (value interface{}, exists bool)
	}); ok {
		if metas, exists := ginCtx.Get(MetaKey); exists {
			return WrapMeta(nil, metas.([]Meta)...)
		}
	}

	metas, ok := ctx.Value(MetaKey).([]Meta)
	if !ok {
		return nil
	}
	return WrapMeta(nil, metas...)
}

// WithFields adds fields to context
func WithFields(ctx context.Context, fields ...Meta) context.Context {
	// Try to set gin context if available (optional dependency)
	if ginCtx, ok := ctx.(interface {
		Get(key string) (value interface{}, exists bool)
		Set(key string, value interface{})
	}); ok {
		if metas, exists := ginCtx.Get(MetaKey); exists {
			metas = append(metas.([]Meta), fields...)
			ginCtx.Set(MetaKey, metas)
		} else {
			ginCtx.Set(MetaKey, fields)
		}
		return ctx
	}

	metas, ok := ctx.Value(MetaKey).([]Meta)
	if !ok {
		ctx = context.WithValue(ctx, MetaKey, fields)
	} else {
		metas = append(metas, fields...)
		ctx = context.WithValue(ctx, MetaKey, metas)
	}
	return ctx
}

// SlogAttrs converts Meta slice to slog.Attr slice
func SlogAttrs(metas []Meta) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(metas))
	for _, meta := range metas {
		attrs = append(attrs, slog.Any(meta.Key(), meta.Value()))
	}
	return attrs
}

// SlogAny converts Meta slice to []any for slog methods
func SlogAny(metas []Meta) []any {
	fields := make([]any, 0, len(metas)*2)
	for _, meta := range metas {
		fields = append(fields, meta.Key(), meta.Value())
	}
	return fields
}
