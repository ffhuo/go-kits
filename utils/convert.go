package utils

import (
	"fmt"
	"reflect"
	"unsafe"
)

// BytesToString 没有内存开销的转换
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes 没有内存开销的转换
func StringToBytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

// ToMap 将结构体转为单层map
func ToMap(in interface{}) (map[string]interface{}, error) {
	// 当前函数只接收struct类型
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr { // 结构体指针
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ToMap only accepts struct or struct pointer; got %T", v)
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
		return fmt.Errorf("ToStruct only accepts struct pointer")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("ToStruct only accepts struct")
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
