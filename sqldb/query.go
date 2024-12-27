package sqldb

import "gorm.io/gorm"

// Args 查询参数构建器
type Args func(*gorm.DB) *gorm.DB

// Where 构建 WHERE 条件
func Where(query interface{}, args ...interface{}) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Where(query, args...)
	}
}

// Order 构建排序条件
func Order(query interface{}) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Order(query)
	}
}

// Unscoped 取消软删除过滤
func Unscoped() Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Unscoped()
	}
}

// WhereIn 构建 IN 条件
func WhereIn(query string, arg interface{}) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Where(query+" in (?)", arg)
	}
}

// WhereBetween 构建 BETWEEN 条件
func WhereBetween(query string, args ...interface{}) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Where(query+" between ? and ?", args...)
	}
}

// WhereNotBetween 构建 NOT BETWEEN 条件
func WhereNotBetween(query string, args ...interface{}) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Where(query+" not between ? and ?", args...)
	}
}

// LeftJoin 构建 LEFT JOIN 条件
func LeftJoin(table, column1, column2 string) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Joins("LEFT JOIN "+table+" ON "+column1+" = "+column2)
	}
}

// Preload 预加载关联
func Preload(query string, args ...interface{}) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Preload(query, args...)
	}
}

// Select 选择字段
func Select(query interface{}, args ...interface{}) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Select(query, args...)
	}
}

// Group 分组
func Group(query string) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Group(query)
	}
}

// Having 分组条件
func Having(query interface{}, args ...interface{}) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Having(query, args...)
	}
}

// Limit 限制数量
func Limit(limit int) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Limit(limit)
	}
}

// Offset 偏移量
func Offset(offset int) Args {
	return func(c *gorm.DB) *gorm.DB {
		return c.Offset(offset)
	}
}
