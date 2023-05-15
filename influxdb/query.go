package influx

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
)

type Args func(*string) *string

func Yield() Args {
	return func(query *string) *string {
		*query += " |> yield()"
		return query
	}
}

func Range(start, stop string) Args {
	return func(query *string) *string {
		template := " |> range("
		if start != "" {
			template += "start:{start}"
			if stop != "" {
				template += ","
			}
		}
		if stop != "" {
			template += "stop:{stop}"
		}
		template += ")"

		if start != "" {
			template = strings.Replace(template, "{start}", start, -1)
		}
		if stop != "" {
			template = strings.Replace(template, "{stop}", stop, -1)
		}

		*query = *query + template
		return query
	}
}

func RangeTime(start, stop *time.Time) Args {
	return func(query *string) *string {
		template := " |> range("
		if start != nil {
			template += "start:{start}"
			if stop != nil {
				template += ","
			}
		}
		if stop != nil {
			template += "stop:{stop}"
		}
		template += ")"

		if start != nil {
			template = strings.Replace(template, "{start}", start.Format("2006-01-02T15:04:05Z"), -1)
		}
		if stop != nil {
			template = strings.Replace(template, "{stop}", stop.Format("2006-01-02T15:04:05Z"), -1)
		}
		*query = *query + template
		return query
	}
}

func From(bucket string) Args {
	return func(query *string) *string {
		template := fmt.Sprintf(" from(bucket:\"%s\") ", bucket)
		*query = *query + template
		return query
	}
}

func Measurement(measurement string) Args {
	return func(query *string) *string {
		template := fmt.Sprintf(" |> filter(fn: (r) => r[\"_measurement\"] == \"%s\") ", measurement)
		*query = *query + template
		return query
	}
}

func FilterField(field string) Args {
	return func(query *string) *string {
		template := fmt.Sprintf(" |> filter(fn: (r) => r[\"_field\"] == \"%s\") ", field)
		*query = *query + template
		return query
	}
}

func Contains(field string, args []string) Args {
	return func(query *string) *string {
		if len(args) == 0 {
			return query
		}
		template := fmt.Sprintf(" |> filter(fn: (r) => contains(value:r.%s, set: [{value}])) ", field)
		value := ""
		for _, v := range args {
			value += fmt.Sprintf(",\"%s\"", v)
		}
		value = value[1:]
		*query = *query + strings.ReplaceAll(template, "{value}", value)
		return query
	}
}

func Group(columns ...string) Args {
	return func(query *string) *string {
		if len(columns) == 0 {
			return query
		}
		cols := ""
		for _, col := range columns {
			cols += ",\"" + col + "\""
		}
		cols = cols[1:]
		*query = *query + fmt.Sprintf(" |> group(columns: [%s])", cols)
		return query
	}
}

func Window(str string) Args {
	return func(query *string) *string {
		*query = *query + fmt.Sprintf(" |> window(every: %s)", str)
		return query
	}
}

func Avg() Args {
	return func(query *string) *string {
		*query = *query + " |> mean()"
		return query
	}
}

func Sum(field string) Args {
	return func(query *string) *string {
		if field != "" {
			*query = *query + fmt.Sprintf(" |> sum(\"%s\")", field)
		} else {
			*query = *query + " |> sum()"
		}
		return query
	}
}

func AggregateWindow(timeRange string, fn string) Args {
	return func(query *string) *string {
		*query = *query + fmt.Sprintf(" |> aggregateWindow(every: %s, fn: %s, createEmpty: false) ", timeRange, fn)
		return query
	}
}

func CumulativeSum() Args {
	return func(query *string) *string {
		*query = *query + " |> cumulativeSum()"
		return query
	}
}

func Eq(field string, arg interface{}) Args {
	return FilterTag(field, "==", arg)
}

func Lt(field string, arg interface{}) Args {
	return FilterTag(field, "<", arg)
}

func Lte(field string, arg interface{}) Args {
	return FilterTag(field, "<=", arg)
}

func Gt(field string, arg interface{}) Args {
	return FilterTag(field, ">", arg)
}

func Gte(field string, arg interface{}) Args {
	return FilterTag(field, ">=", arg)
}

func Ne(field string, arg interface{}) Args {
	return FilterTag(field, "!=", arg)
}

func FilterTag(field, cond string, arg interface{}) Args {
	return func(query *string) *string {
		template := fmt.Sprintf(" |> filter(fn: (r) => r[\"%s\"] %s \"%s\")", field, cond, arg)
		*query = *query + template
		return query
	}
}

func FilterContains(field string, args ...string) Args {
	return func(query *string) *string {
		if len(args) == 0 {
			return query
		}
		template := " |> filter(fn: (r) => %s)"
		var cond string
		for _, arg := range args {
			if cond != "" {
				cond += " or "
			}
			cond += fmt.Sprintf("r[\"%v\"] == \"%v\"", field, arg)
		}

		*query = *query + fmt.Sprintf(template, cond)
		return query
	}
}

func FilterNotContains(field string, args ...string) Args {
	return func(query *string) *string {
		if len(args) == 0 {
			return query
		}
		template := " |> filter(fn: (r) => %s)"
		var cond string
		for _, arg := range args {
			if cond != "" {
				cond += " and "
			}
			cond += fmt.Sprintf("r[\"%v\"] != \"%v\"", field, arg)
		}

		*query = *query + fmt.Sprintf(template, cond)
		return query
	}
}

func FilterFluxTag(str string, args ...string) Args {
	return func(query *string) *string {
		template := fmt.Sprintf(" |> filter(fn: (r) => %s)", str)
		for _, v := range args {
			template = strings.Replace(template, "?", "\""+v+"\"", 1)
		}
		*query = *query + template
		return query
	}
}

func Scan(result *api.QueryTableResult, dest interface{}) error {
	switch dt := dest.(type) {
	case map[string]interface{}, *map[string]interface{}:
		mapValue, ok := dt.(map[string]interface{})
		if !ok {
			if v, ok := dt.(*map[string]interface{}); ok {
				if *v == nil {
					*v = map[string]interface{}{}
				}
				mapValue = *v
			}
		}
		for result.Next() {
			for k, v := range result.Record().Values() {
				mapValue[k] = v
			}
		}
	case *[]map[string]interface{}:
		for result.Next() {
			mapValue := map[string]interface{}{}
			for k, v := range result.Record().Values() {
				mapValue[k] = v
			}
			*dt = append(*dt, mapValue)
		}
	case *int, *int8, *int16, *int32, *int64,
		*uint, *uint8, *uint16, *uint32, *uint64, *uintptr,
		*float32, *float64,
		*bool, *string, *time.Time:
	default:
		destValue := reflect.ValueOf(dest)
		if destValue.Kind() == reflect.Ptr {
			if destValue.IsNil() && destValue.CanAddr() {
				destValue.Set(reflect.New(destValue.Type().Elem()))
			}
			destValue = destValue.Elem()
		}
		reflectValue := destValue
		if reflectValue.Kind() == reflect.Interface {
			reflectValue = reflectValue.Elem()
		}

		reflectValueType := reflectValue.Type()
		switch reflectValueType.Kind() {
		case reflect.Array, reflect.Slice:
			reflectValueType = reflectValueType.Elem()
		}

		isPtr := reflectValueType.Kind() == reflect.Ptr
		if isPtr {
			reflectValueType = reflectValueType.Elem()
		}
		switch reflectValue.Kind() {
		case reflect.Array, reflect.Slice:
			var (
				index       int
				isArrayKind = reflectValue.Kind() == reflect.Array
			)
			if reflectValue.Cap() == 0 {
				destValue.Set(reflect.MakeSlice(reflectValue.Type(), 0, 20))
			} else if !isArrayKind {
				reflectValue.SetLen(0)
				destValue.Set(reflectValue)
			}

			for result.Next() {
				elem := reflect.New(reflectValueType).Elem()
				if err := scanToStruct(result.Record().Values(), elem); err != nil {
					return err
				}
				if !isArrayKind {
					reflectValue = reflect.Append(reflectValue, elem)
				} else {
					reflectValue.Index(index).Set(elem)
				}
				index++
			}
			destValue.Set(reflectValue)
		case reflect.Struct, reflect.Ptr:
			for result.Next() {
				return scanToStruct(result.Record().Values(), reflectValue)
			}
		}
	}
	return nil
}

func scanToStruct(data map[string]interface{}, dest reflect.Value) error {
	// v := reflect.ValueOf(dest)
	if dest.Kind() == reflect.Ptr {
		dest = dest.Elem()
	}
	if dest.Kind() != reflect.Struct {
		return errors.New("scanToStruct only accepts struct now: " + dest.Kind().String())
	}
	findFromMap := func(key string, nameTag string) interface{} {
		for k, v := range data {
			if k == key || k == nameTag {
				return v
			}
		}
		return nil
	}
	for i := 0; i < dest.NumField(); i++ {
		val := findFromMap(dest.Type().Field(i).Name, dest.Type().Field(i).Tag.Get("json"))
		if val != nil { //&& reflect.ValueOf(val).Kind() == v.Field(i).Kind()
			dest.Field(i).Set(reflect.ValueOf(val))
		}
	}
	return nil
}

// func parse(v reflect.Value) (tags map[string]string, fields map[string]interface{}, err error) {
// 	if v.Kind() == reflect.Ptr { // 结构体指针
// 		v = v.Elem()
// 	}
// 	if v.Kind() != reflect.Struct {
// 		err = errors.New("parse only accepts struct or struct pointer")
// 		return
// 	}

// 	queue := make([]interface{}, 0, 1)
// 	queue = append(queue, v.Interface())

// 	parseTag := func(ti reflect.StructField) (t, k string) {
// 		t = ti.Tag.Get("type")
// 		k = ti.Tag.Get("json")
// 		return
// 	}

// 	tags = map[string]string{}
// 	fields = map[string]interface{}{}

// 	for len(queue) > 0 {
// 		v := reflect.ValueOf(queue[0])
// 		if v.Kind() == reflect.Ptr { // 结构体指针
// 			v = v.Elem()
// 		}
// 		queue = queue[1:]
// 		t := v.Type()
// 		for i := 0; i < v.NumField(); i++ {
// 			vi := v.Field(i)
// 			if vi.Kind() == reflect.Ptr { // 内嵌指针
// 				vi = vi.Elem()
// 				if vi.Kind() == reflect.Struct { // 结构体
// 					queue = append(queue, vi.Interface())
// 				} else {
// 					t, k := parseTag(t.Field(i))
// 					if k != "" && k != "-" {
// 						if t == "tag" {
// 							tags[k] = vi.String()
// 						} else {
// 							fields[k] = vi.Interface()
// 						}
// 					}
// 				}
// 				break
// 			}
// 			if vi.Kind() == reflect.Struct { // 内嵌结构体
// 				queue = append(queue, vi.Interface())
// 				break
// 			}
// 			// 一般字段
// 			t, k := parseTag(t.Field(i))
// 			if k != "" && k != "-" {
// 				if t == "tag" {
// 					tags[k] = vi.String()
// 				} else {
// 					fields[k] = vi.Interface()
// 				}
// 			}
// 		}
// 	}
// 	return
// }
