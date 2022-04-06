package logger

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var MetaKey = "loggerMeta"
var _ Meta = (*meta)(nil)

// Meta key-value
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

// NewMeta create meat
func NewMeta(key string, value interface{}) Meta {
	return &meta{key: key, value: value}
}

// WrapMeta wrap meta to zap fields
func WrapMeta(err error, metas ...Meta) (fields []zap.Field) {
	capacity := len(metas) + 1 // namespace meta
	if err != nil {
		capacity++
	}

	fields = make([]zap.Field, 0, capacity)
	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	fields = append(fields, zap.Namespace("meta"))
	for _, meta := range metas {
		fields = append(fields, zap.Any(meta.Key(), meta.Value()))
	}

	return
}

func Fields(ctx context.Context) []zap.Field {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		metas, ok := ginCtx.Get(MetaKey)
		if !ok {
			return nil
		}
		return WrapMeta(nil, metas.([]Meta)...)
	}

	metas, ok := ctx.Value(MetaKey).([]Meta)
	if !ok {
		return nil
	}
	return WrapMeta(nil, metas...)
}

func WithFields(ctx context.Context, fields ...Meta) context.Context {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		metas, ok := ginCtx.Get(MetaKey)
		if ok {
			metas = append(metas.([]Meta), fields...)
		} else {
			metas = fields
		}
		ginCtx.Set(MetaKey, metas)
		return ginCtx
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
