package field_test

import (
	"context"
	"fmt"

	"github.com/ffhuo/go-kits/common/field"
)

func ExampleF() {
	// 创建字段
	userID := field.F("user_id", "12345")
	fmt.Printf("Key: %s, Value: %v\n", userID.Key(), userID.Value())
	// Output: Key: user_id, Value: 12345
}

func ExampleWith() {
	ctx := context.Background()

	// 添加字段到context
	ctx = field.With(ctx,
		field.F("user_id", "12345"),
		field.F("request_id", "req-abc"),
	)

	// 获取字段
	fields := field.Get(ctx)
	fmt.Printf("Fields count: %d\n", len(fields))
	// Output: Fields count: 2
}

func ExampleGet() {
	ctx := context.Background()
	ctx = field.With(ctx,
		field.F("service", "user"),
		field.F("action", "login"),
	)

	// 获取所有字段
	fields := field.Get(ctx)
	for _, f := range fields {
		fmt.Printf("%s: %v\n", f.Key(), f.Value())
	}
	// Output:
	// service: user
	// action: login
}

// 模拟中间件使用
func ExampleWith_middleware() {
	ctx := context.Background()

	// 模拟gin中间件添加请求字段
	ctx = field.With(ctx,
		field.F("ip", "192.168.1.1"),
		field.F("method", "GET"),
		field.F("path", "/api/users"),
	)

	// 模拟业务逻辑添加更多字段
	ctx = field.With(ctx,
		field.F("user_id", "12345"),
		field.F("action", "list_users"),
	)

	// 获取所有字段
	fields := field.Get(ctx)
	fmt.Printf("Total fields: %d\n", len(fields))
	// Output: Total fields: 5
}
