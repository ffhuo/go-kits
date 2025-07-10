package logger_v2

import (
	"context"
	"log/slog"

	"github.com/ffhuo/go-kits/common/field"
)

// 重新导出，保持兼容性
type Meta = field.Meta

var F = field.F
var With = field.With

// 向后兼容的函数
var WithFields = field.With

func NewMeta(key string, value interface{}) Meta { return field.F(key, value) }

const MetaKey = "fields"

// Fields 从context获取字段并转换为slog格式
func Fields(ctx context.Context) []any {
	metas := field.Get(ctx)
	if len(metas) == 0 {
		return nil
	}

	fields := make([]any, 0, len(metas)*2)
	for _, meta := range metas {
		fields = append(fields, meta.Key(), meta.Value())
	}
	return fields
}

// SlogAttrs 将Meta切片转换为slog.Attr切片
func SlogAttrs(metas []Meta) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(metas))
	for _, meta := range metas {
		attrs = append(attrs, slog.Any(meta.Key(), meta.Value()))
	}
	return attrs
}

// SlogAny 将Meta切片转换为[]any用于slog方法
func SlogAny(metas []Meta) []any {
	fields := make([]any, 0, len(metas)*2)
	for _, meta := range metas {
		fields = append(fields, meta.Key(), meta.Value())
	}
	return fields
}
