package field

import (
	"context"
	"testing"
)

func TestF(t *testing.T) {
	meta := F("test_key", "test_value")
	if meta.Key() != "test_key" {
		t.Errorf("Expected key 'test_key', got '%s'", meta.Key())
	}
	if meta.Value() != "test_value" {
		t.Errorf("Expected value 'test_value', got '%v'", meta.Value())
	}
}

func TestWith(t *testing.T) {
	ctx := context.Background()

	// 测试添加字段
	ctx = With(ctx, F("key1", "value1"))
	metas := Get(ctx)
	if len(metas) != 1 {
		t.Errorf("Expected 1 meta, got %d", len(metas))
	}
	if metas[0].Key() != "key1" || metas[0].Value() != "value1" {
		t.Errorf("Expected key1=value1, got %s=%v", metas[0].Key(), metas[0].Value())
	}

	// 测试添加多个字段
	ctx = With(ctx, F("key2", "value2"), F("key3", 123))
	metas = Get(ctx)
	if len(metas) != 3 {
		t.Errorf("Expected 3 metas, got %d", len(metas))
	}
}

func TestGet(t *testing.T) {
	ctx := context.Background()

	// 测试空context
	metas := Get(ctx)
	if metas != nil {
		t.Errorf("Expected nil metas for empty context, got %v", metas)
	}

	// 测试有字段的context
	ctx = With(ctx, F("test", "value"))
	metas = Get(ctx)
	if len(metas) != 1 {
		t.Errorf("Expected 1 meta, got %d", len(metas))
	}
}

// 模拟gin.Context的接口
type mockGinContext struct {
	context.Context
	data map[string]interface{}
}

func newMockGinContext() *mockGinContext {
	return &mockGinContext{
		Context: context.Background(),
		data:    make(map[string]interface{}),
	}
}

func (m *mockGinContext) Get(key string) (value interface{}, exists bool) {
	if m.data == nil {
		return nil, false
	}
	value, exists = m.data[key]
	return
}

func (m *mockGinContext) Set(key string, value interface{}) {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
}

func TestWithGinContext(t *testing.T) {
	ctx := newMockGinContext()

	// 测试添加字段到gin context
	result := With(ctx, F("gin_key", "gin_value"))
	if result != ctx {
		t.Error("Expected same context to be returned")
	}

	// 验证字段是否正确存储
	metas := Get(ctx)
	if len(metas) != 1 {
		t.Errorf("Expected 1 meta, got %d", len(metas))
	}
	if metas[0].Key() != "gin_key" || metas[0].Value() != "gin_value" {
		t.Errorf("Expected gin_key=gin_value, got %s=%v", metas[0].Key(), metas[0].Value())
	}

	// 测试添加更多字段
	With(ctx, F("gin_key2", "gin_value2"))
	metas = Get(ctx)
	if len(metas) != 2 {
		t.Errorf("Expected 2 metas, got %d", len(metas))
	}
}
