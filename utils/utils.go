package utils

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// 驼峰式写法转为下划线写法
func CamelToCase(name string) string {
	var (
		value                          = name
		buf                            strings.Builder
		lastCase, nextCase, nextNumber bool // upper case == true
		curCase                        = value[0] <= 'Z' && value[0] >= 'A'
	)

	for i, v := range value[:len(value)-1] {
		nextCase = value[i+1] <= 'Z' && value[i+1] >= 'A'
		nextNumber = value[i+1] >= '0' && value[i+1] <= '9'

		if curCase {
			if lastCase && (nextCase || nextNumber) {
				buf.WriteRune(v + 32)
			} else {
				if i > 0 && value[i-1] != '_' && value[i+1] != '_' {
					buf.WriteByte('_')
				}
				buf.WriteRune(v + 32)
			}
		} else {
			buf.WriteRune(v)
		}

		lastCase = curCase
		curCase = nextCase
	}

	if curCase {
		if !lastCase && len(value) > 1 {
			buf.WriteByte('_')
		}
		buf.WriteByte(value[len(value)-1] + 32)
	} else {
		buf.WriteByte(value[len(value)-1])
	}
	ret := buf.String()
	return ret
}

// ToMap 将结构体转为单层map
func ToMap(in interface{}) (map[string]interface{}, error) {
	// 当前函数只接收struct类型
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr { // 结构体指针
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.Errorf("ToMap only accepts struct or struct pointer; got %T", v)
	}

	out := make(map[string]interface{})
	queue := make([]interface{}, 0, 1)
	queue = append(queue, in)

	for len(queue) > 0 {
		v := reflect.ValueOf(queue[0])
		if v.Kind() == reflect.Ptr { // 结构体指针
			v = v.Elem()
		}
		queue = queue[1:]
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			vi := v.Field(i)
			if vi.Kind() == reflect.Ptr { // 内嵌指针
				vi = vi.Elem()
				if vi.Kind() == reflect.Struct { // 结构体
					queue = append(queue, vi.Interface())
				} else {
					ti := t.Field(i)
					if tagValue := ti.Tag.Get("json"); tagValue != "" && tagValue != "-" {
						// 存入map
						out[tagValue] = vi.Interface()
					}
				}
				break
			}
			if vi.Kind() == reflect.Struct { // 内嵌结构体
				queue = append(queue, vi.Interface())
				break
			}
			// 一般字段
			ti := t.Field(i)
			if tagValue := ti.Tag.Get("json"); tagValue != "" && tagValue != "-" {
				// 存入map
				out[tagValue] = vi.Interface()
			}
		}
	}
	return out, nil
}

func ToStruct(m map[string]interface{}, u interface{}) error {
	v := reflect.ValueOf(u)
	if v.Kind() != reflect.Ptr {
		return errors.New("ToStruct only accepts struct pointer")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("ToStruct only accepts struct")
	}
	findFromMap := func(key string, nameTag string) interface{} {
		for k, v := range m {
			if k == key || k == nameTag {
				return v
			}
		}
		return nil
	}
	for i := 0; i < v.NumField(); i++ {
		val := findFromMap(v.Type().Field(i).Name, v.Type().Field(i).Tag.Get("json"))
		if val != nil { //&& reflect.ValueOf(val).Kind() == v.Field(i).Kind()
			v.Field(i).Set(reflect.ValueOf(val))
		}
	}
	return nil
}
