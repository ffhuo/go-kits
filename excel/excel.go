// Package excel provides simplified functions for working with Excel files
package excel

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Excel represents a wrapper around excelize.File with simplified methods
type Excel struct {
	*excelize.File
	ActiveSheet string
}

// New creates a new Excel file with the specified sheet name
// If sheetName is empty, it will use the default "Sheet1"
func New(sheetName string) *Excel {
	f := excelize.NewFile()
	activeSheet := "Sheet1"

	if sheetName != "" {
		sheet, _ := f.NewSheet(sheetName)
		f.DeleteSheet("Sheet1")
		f.SetActiveSheet(sheet)
		activeSheet = sheetName
	}

	return &Excel{
		File:        f,
		ActiveSheet: activeSheet,
	}
}

// OpenReader opens an Excel file from an io.Reader
func OpenReader(r io.Reader) (*Excel, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	return &Excel{
		File:        f,
		ActiveSheet: sheets[0],
	}, nil
}

// SetSheet sets the active sheet
// Returns error if sheet doesn't exist
func (e *Excel) SetSheet(sheetName string) error {
	sheets := e.GetSheetList()
	for _, s := range sheets {
		if s == sheetName {
			e.ActiveSheet = sheetName
			return nil
		}
	}
	return fmt.Errorf("sheet '%s' not found", sheetName)
}

// WriteTable writes a simple table with headers and data to the active sheet
// headers: slice of column headers
// data: 2D slice of data to write below the headers
func (e *Excel) WriteTable(headers []string, data [][]interface{}) error {
	// Write headers
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := e.SetCellValue(e.ActiveSheet, cell, header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Write data
	for row, rowData := range data {
		for col, value := range rowData {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2) // +2 because headers are at row 1
			if err := e.SetCellValue(e.ActiveSheet, cell, value); err != nil {
				return fmt.Errorf("failed to write data at cell %s: %w", cell, err)
			}
		}
	}

	return nil
}

// WriteStructs writes a slice of structs to the active sheet
// The struct should have `xlsx:"column_name"` tags to specify column headers
// Example:
//
//	type Person struct {
//	    Name string `xlsx:"姓名"`
//	    Age  int    `xlsx:"年龄"`
//	}
//
// skipFields: optional list of field names to skip
func (e *Excel) WriteStructs(data interface{}, skipFields ...string) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return fmt.Errorf("data must be a slice or array, got %v", v.Kind())
	}

	if v.Len() == 0 {
		return nil // Nothing to write
	}

	// Get the first element to extract field information
	firstItem := v.Index(0)
	if firstItem.Kind() == reflect.Ptr {
		firstItem = firstItem.Elem()
	}

	if firstItem.Kind() != reflect.Struct {
		return fmt.Errorf("slice elements must be structs, got %v", firstItem.Kind())
	}

	// Create a map of fields to skip
	skipMap := make(map[string]bool)
	for _, field := range skipFields {
		skipMap[field] = true
	}

	// Extract headers and field indices
	headers := []string{}
	fieldIndices := []int{}
	t := firstItem.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		header := field.Tag.Get("xlsx")

		// Skip fields with no tag, "-" tag, or in skip list
		if header == "" || header == "-" || skipMap[header] {
			continue
		}

		headers = append(headers, header)
		fieldIndices = append(fieldIndices, i)
	}

	// Write headers
	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		if err := e.SetCellValue(e.ActiveSheet, cell, header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Write data
	for row := 0; row < v.Len(); row++ {
		item := v.Index(row)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		for col, fieldIdx := range fieldIndices {
			field := item.Field(fieldIdx)
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2) // +2 because headers are at row 1

			// Set cell value based on field type
			var err error
			switch field.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				err = e.SetCellValue(e.ActiveSheet, cell, field.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				err = e.SetCellValue(e.ActiveSheet, cell, field.Uint())
			case reflect.Float32, reflect.Float64:
				err = e.SetCellValue(e.ActiveSheet, cell, field.Float())
			case reflect.String:
				err = e.SetCellValue(e.ActiveSheet, cell, field.String())
			case reflect.Bool:
				err = e.SetCellValue(e.ActiveSheet, cell, field.Bool())
			default:
				// Try to use the default string representation
				err = e.SetCellValue(e.ActiveSheet, cell, field.Interface())
			}

			if err != nil {
				return fmt.Errorf("failed to write data at cell %s: %w", cell, err)
			}
		}
	}

	return nil
}

// ReadToMaps reads data from the active sheet and returns it as a slice of maps
// Each map represents a row with column headers as keys and cell values as values
func (e *Excel) ReadToMaps() ([]map[string]string, error) {
	rows, err := e.Rows(e.ActiveSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	var result []map[string]string
	var headers []string

	// Read headers
	if rows.Next() {
		headers, err = rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to read header row: %w", err)
		}
	} else {
		return nil, fmt.Errorf("empty sheet")
	}

	// Read data rows
	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to read data row: %w", err)
		}

		if len(columns) == 0 {
			continue
		}

		row := make(map[string]string)
		for i, value := range columns {
			if i < len(headers) {
				row[headers[i]] = strings.TrimSpace(value)
			}
		}
		result = append(result, row)
	}

	return result, nil
}

// ReadSheet reads data from the specified sheet and returns it as a slice of maps
// If sheet is empty, uses the active sheet
func (e *Excel) ReadSheet(sheet string) ([]map[string]string, error) {
	if sheet != "" {
		if err := e.SetSheet(sheet); err != nil {
			return nil, err
		}
	}
	return e.ReadToMaps()
}

// Read is a convenience function that opens an Excel file from an io.Reader
// and reads data from the specified sheet
// If sheet is empty, reads from the first sheet
func Read(file io.Reader, sheet string) ([]map[string]string, error) {
	excel, err := OpenReader(file)
	if err != nil {
		return nil, err
	}

	if sheet != "" {
		if err := excel.SetSheet(sheet); err != nil {
			return nil, err
		}
	}

	return excel.ReadToMaps()
}
