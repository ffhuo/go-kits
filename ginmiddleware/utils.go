package ginmiddleware

import "strings"

// contains 检查字符串切片是否包含指定字符串（精确匹配）
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// containsIgnoreCase 检查字符串切片是否包含指定字符串（忽略大小写）
func containsIgnoreCase(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// isPrintableText 检查字符串是否为可打印文本
func isPrintableText(s string) bool {
	for _, r := range s {
		if r < 32 && r != 9 && r != 10 && r != 13 { // 除了tab、换行、回车外的控制字符
			return false
		}
	}
	return true
}

// maskSensitiveData 脱敏敏感数据
func maskSensitiveData(data string, sensitiveFields []string) string {
	result := data
	for _, field := range sensitiveFields {
		// 简单的脱敏处理，将敏感字段的值替换为***
		patterns := []string{
			field + "=",
			"\"" + field + "\":",
			"'" + field + "':",
		}

		for _, pattern := range patterns {
			if idx := strings.Index(strings.ToLower(result), strings.ToLower(pattern)); idx != -1 {
				start := idx + len(pattern)
				end := start

				// 查找值的结束位置
				if start < len(result) {
					char := result[start]
					if char == '"' || char == '\'' {
						// JSON或带引号的值
						quote := char
						end = start + 1
						for end < len(result) && result[end] != byte(quote) {
							end++
						}
						if end < len(result) {
							// 保留引号，只替换引号内的内容
							result = result[:start+1] + "***" + result[end:]
						}
					} else {
						// URL参数或普通值
						for end < len(result) && result[end] != '&' && result[end] != ' ' && result[end] != ',' && result[end] != '}' {
							end++
						}
						if end > start {
							result = result[:start] + "***" + result[end:]
						}
					}
				}
			}
		}
	}
	return result
}
