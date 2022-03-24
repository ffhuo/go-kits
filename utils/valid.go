package utils

import (
	"regexp"
)

func ValidEamil(email string) bool {
	emailRegexp := "^[A-Za-z0-9]+([-_.][A-Za-z0-9]+)*@([A-Za-z0-9]+[-.])+[A-Za-z0-9]{2,4}$"
	b, err := regexp.MatchString(emailRegexp, email)
	if err != nil {
		return false
	}
	return b
}

func ValidMobile(phone string) bool {
	reg := `^1([38][0-9]|14[579]|5[^4]|16[6]|7[1-35-8]|9[189])\d{8}$`
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(phone)
}

// 敏感数据保护-手机号码保护：仅展示前3个数字和后2个数字，其他用“*”保护，比如：186******48
func ProtectMobile(mobile string) string {
	// 如果手机号为空字符串，那么直接返回
	if mobile == "" {
		return mobile
	}
	return mobile[:3] + "******" + mobile[len(mobile)-2:]
}
