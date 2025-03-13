package utils

import (
	"reflect"
)

func Compare(a interface{}, b interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)
	aType := reflect.TypeOf(a)
	bType := reflect.TypeOf(b)
	for i := 0; i < aType.Elem().NumField(); i++ {
		at := aType.Elem().Field(i)
		j := 0
		for ; j < bType.Elem().NumField(); j++ {
			if bType.Elem().Field(j).Name == at.Name {
				break
			}
		}
		if j >= bType.Elem().NumField() {
			continue
		}
		av := aValue.Elem().Field(i).Interface()
		bv := bValue.Elem().Field(j).Interface()
		if av != bv {
			result[CamelToCase(at.Name)] = bv
		}
	}
	return result
}
