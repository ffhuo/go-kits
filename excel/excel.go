package excel

import (
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func NewExcel(sheetName string) *excelize.File {
	f := excelize.NewFile()
	f.SetDefaultFont("Arial")
	if sheetName != "" {
		sheet, _ := f.NewSheet(sheetName)
		f.DeleteSheet("Sheet1")
		f.SetActiveSheet(sheet)
	}
	return f
}

func getCellNum(row, column int) string {
	var result string
	start := 'A'
	index := row / 26
	left := row % 26
	if index > 0 {
		result = string(start + rune(index-1))
	}

	result += string(start + rune(left))
	result += strconv.Itoa(column)

	return result
}

func WriteToXlsx(f *excelize.File, sheetName string, title []string, data [][]interface{}) error {
	for i, field := range title {
		f.SetCellValue(sheetName, getCellNum(i, 1), field)
	}

	for i, row := range data {
		for index, value := range row {
			f.SetCellValue(sheetName, getCellNum(index, i+2), value)
			// tempValue := reflect.ValueOf(value)
			// switch tempValue.Kind() {
			// case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8:
			// 	f.SetCellValue(sheetName, getCellNum(index, i+2), tempValue.Int())
			// case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint8, reflect.Uint64:
			// 	f.SetCellValue(sheetName, getCellNum(index, i+2), tempValue.Uint())
			// case reflect.String:
			// 	f.SetCellValue(sheetName, getCellNum(index, i+2), tempValue.String())
			// case reflect.Float32, reflect.Float64:
			// 	f.SetCellValue(sheetName, getCellNum(index, i+2), tempValue.Float())
			// default:
			// 	return fmt.Errorf("write file type unKnow: %v", tempValue.Kind())
			// }
		}
	}
	return nil
}

// WriteDataToXlsx ...
func WriteDataToXlsx(f *excelize.File, sheetName string, data interface{}, example interface{}, excludeField ...string) error {
	cloumns := map[int]bool{}
	rt := reflect.TypeOf(example).Elem()
	index := 0
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		tempValue := field.Tag.Get("xlsx")
		if tempValue == "" || tempValue == "-" {
			continue
		}

		isExclude := false
		for _, field := range excludeField {
			if field == tempValue {
				isExclude = true
				break
			}
		}
		if isExclude {
			continue
		}

		cloumns[i] = true
		f.SetCellValue(sheetName, getCellNum(index, 1), tempValue)
		index++
	}

	dataValues := reflect.ValueOf(data)
	if dataValues.Kind() != reflect.Slice && dataValues.Kind() != reflect.Array {
		return fmt.Errorf("data is not a slice or array")
	}

	for i := 0; i < dataValues.Len(); i++ {
		value := dataValues.Index(i)
		index = 0
		for j := 0; j < value.NumField(); j++ {
			_, ok := cloumns[j]
			if !ok {
				continue
			}
			field := value.Field(j)
			switch field.Kind() {
			case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8:
				f.SetCellValue(sheetName, getCellNum(index, i+2), field.Int())
			case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint8, reflect.Uint64:
				f.SetCellValue(sheetName, getCellNum(index, i+2), field.Uint())
			case reflect.String:
				f.SetCellValue(sheetName, getCellNum(index, i+2), field.String())
			case reflect.Float32, reflect.Float64:
				f.SetCellValue(sheetName, getCellNum(index, i+2), field.Float())
			default:
				fmt.Errorf("write file type unKnow: %v", field.Kind())
			}
			index++
		}
	}
	return nil
}

func Read(file io.Reader, sheet string) ([]map[string]string, error) {
	var (
		err   error
		items []map[string]string = make([]map[string]string, 0)
	)

	xlsx, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel file: %v", err)
	}

	sheets := xlsx.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheet in Excel file")
	}

	if sheet != "" {
		hasSheet := false
		for _, s := range sheets {
			if s == sheet {
				hasSheet = true
				break
			}
		}

		if !hasSheet {
			return nil, fmt.Errorf("sheet %s not found", sheet)
		}
	} else {
		sheet = sheets[0]
	}

	titles := make(map[int]string)
	rows, err := xlsx.Rows(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel row: %v", err)
	}
	// 读取首行
	if rows.Next() {
		title, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to read Excel column: %v", err)
		}
		for i, v := range title {
			titles[i] = v
		}
	}

	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to read Excel column: %v", err)
		}
		if len(row) == 0 {
			break
		}
		item := make(map[string]string, len(row))
		for i, v := range row {
			item[titles[i]] = v
		}
		items = append(items, item)
	}

	return items, nil
}
