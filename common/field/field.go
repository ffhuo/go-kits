package field

import (
	"context"
)

// Meta 日志字段接口
type Meta interface {
	Key() string
	Value() interface{}
}

type meta struct {
	key   string
	value interface{}
}

func (m *meta) Key() string        { return m.key }
func (m *meta) Value() interface{} { return m.value }

// F 创建字段的简化函数
func F(key string, value interface{}) Meta {
	return &meta{key: key, value: value}
}

// With 向context添加字段
func With(ctx context.Context, fields ...Meta) context.Context {
	const metaKey = "fields"

	// 支持gin context
	if ginCtx, ok := ctx.(interface {
		Get(key string) (value interface{}, exists bool)
		Set(key string, value interface{})
	}); ok {
		var metas []Meta
		if existing, exists := ginCtx.Get(metaKey); exists {
			if slice, ok := existing.([]Meta); ok {
				metas = append(slice, fields...)
			} else {
				metas = fields
			}
		} else {
			metas = fields
		}
		ginCtx.Set(metaKey, metas)
		return ctx
	}

	// 普通context
	var metas []Meta
	if existing := ctx.Value(metaKey); existing != nil {
		if slice, ok := existing.([]Meta); ok {
			metas = append(slice, fields...)
		} else {
			metas = fields
		}
	} else {
		metas = fields
	}
	return context.WithValue(ctx, metaKey, metas)
}

// Get 从context获取字段
func Get(ctx context.Context) []Meta {
	const metaKey = "fields"

	// 支持gin context
	if ginCtx, ok := ctx.(interface {
		Get(key string) (value interface{}, exists bool)
	}); ok {
		if metas, exists := ginCtx.Get(metaKey); exists {
			if slice, ok := metas.([]Meta); ok {
				return slice
			}
		}
		return nil
	}

	// 普通context
	if metas := ctx.Value(metaKey); metas != nil {
		if slice, ok := metas.([]Meta); ok {
			return slice
		}
	}
	return nil
}
