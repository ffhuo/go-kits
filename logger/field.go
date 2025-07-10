package logger

import (
	"context"

	"github.com/ffhuo/go-kits/common/field"
	"go.uber.org/zap"
)

// 重新导出，保持兼容性
type Meta = field.Meta

var F = field.F
var With = field.With

// 向后兼容的函数
var WithFields = field.With

func NewMeta(key string, value interface{}) Meta { return field.F(key, value) }

const MetaKey = "fields"

// Fields 从context获取字段并转换为zap.Field
func Fields(ctx context.Context) []zap.Field {
	metas := field.Get(ctx)
	if len(metas) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(metas)+1)
	fields = append(fields, zap.Namespace("meta"))
	for _, meta := range metas {
		fields = append(fields, zap.Any(meta.Key(), meta.Value()))
	}
	return fields
}
