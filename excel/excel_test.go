package excel

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestNew tests the New function
func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		sheetName string
		wantSheet string
	}{
		{
			name:      "default sheet",
			sheetName: "",
			wantSheet: "Sheet1",
		},
		{
			name:      "custom sheet",
			sheetName: "TestSheet",
			wantSheet: "TestSheet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			excel := New(tt.sheetName)
			if excel == nil {
				t.Fatal("Expected excel to not be nil")
			}
			if excel.ActiveSheet != tt.wantSheet {
				t.Errorf("Expected ActiveSheet to be %s, got %s", tt.wantSheet, excel.ActiveSheet)
			}
		})
	}
}

// TestExcel_WriteTable tests the WriteTable method
func TestExcel_WriteTable(t *testing.T) {
	excel := New("TestSheet")

	headers := []string{"Name", "Age", "City"}
	data := [][]interface{}{
		{"John", 30, "New York"},
		{"Alice", 25, "London"},
	}

	err := excel.WriteTable(headers, data)
	if err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}

	// Verify cell values
	val, err := excel.GetCellValue("TestSheet", "A1")
	if err != nil {
		t.Errorf("GetCellValue returned error: %v", err)
	}
	if val != "Name" {
		t.Errorf("Expected cell A1 to be 'Name', got '%s'", val)
	}

	val, err = excel.GetCellValue("TestSheet", "B1")
	if err != nil {
		t.Errorf("GetCellValue returned error: %v", err)
	}
	if val != "Age" {
		t.Errorf("Expected cell B1 to be 'Age', got '%s'", val)
	}

	val, err = excel.GetCellValue("TestSheet", "A2")
	if err != nil {
		t.Errorf("GetCellValue returned error: %v", err)
	}
	if val != "John" {
		t.Errorf("Expected cell A2 to be 'John', got '%s'", val)
	}

	val, err = excel.GetCellValue("TestSheet", "B3")
	if err != nil {
		t.Errorf("GetCellValue returned error: %v", err)
	}
	if val != "25" {
		t.Errorf("Expected cell B3 to be '25', got '%s'", val)
	}
}

// TestExcel_WriteStructs tests the WriteStructs method
func TestExcel_WriteStructs(t *testing.T) {
	type Person struct {
		Name string `xlsx:"Name"`
		Age  int    `xlsx:"Age"`
		City string `xlsx:"City"`
		Skip string `xlsx:"-"`
	}

	people := []Person{
		{Name: "John", Age: 30, City: "New York", Skip: "Hidden"},
		{Name: "Alice", Age: 25, City: "London", Skip: "Hidden"},
	}

	excel := New("TestSheet")

	err := excel.WriteStructs(people)
	if err != nil {
		t.Fatalf("WriteStructs returned error: %v", err)
	}

	// Verify cell values
	val, err := excel.GetCellValue("TestSheet", "A1")
	if err != nil {
		t.Errorf("GetCellValue returned error: %v", err)
	}
	if val != "Name" {
		t.Errorf("Expected cell A1 to be 'Name', got '%s'", val)
	}

	val, err = excel.GetCellValue("TestSheet", "B1")
	if err != nil {
		t.Errorf("GetCellValue returned error: %v", err)
	}
	if val != "Age" {
		t.Errorf("Expected cell B1 to be 'Age', got '%s'", val)
	}

	val, err = excel.GetCellValue("TestSheet", "A2")
	if err != nil {
		t.Errorf("GetCellValue returned error: %v", err)
	}
	if val != "John" {
		t.Errorf("Expected cell A2 to be 'John', got '%s'", val)
	}

	// Verify Skip field is not present
	val, err = excel.GetCellValue("TestSheet", "D1")
	if err != nil {
		t.Errorf("GetCellValue returned error: %v", err)
	}
	if val != "" {
		t.Errorf("Expected cell D1 to be empty, got '%s'", val)
	}
}

// TestRead tests the Read function
func TestRead(t *testing.T) {
	// Create a test Excel file
	f := excelize.NewFile()
	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Age")
	f.SetCellValue("Sheet1", "A2", "John")
	f.SetCellValue("Sheet1", "B2", "30")
	f.SetCellValue("Sheet1", "A3", "Alice")
	f.SetCellValue("Sheet1", "B3", "25")

	// Save to buffer
	buf := new(bytes.Buffer)
	err := f.Write(buf)
	if err != nil {
		t.Fatalf("Failed to write Excel to buffer: %v", err)
	}

	// Read from buffer
	data, err := Read(bytes.NewReader(buf.Bytes()), "")
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	// Verify data
	if len(data) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(data))
	}

	if data[0]["Name"] != "John" {
		t.Errorf("Expected Name to be 'John', got '%s'", data[0]["Name"])
	}
	if data[0]["Age"] != "30" {
		t.Errorf("Expected Age to be '30', got '%s'", data[0]["Age"])
	}
	if data[1]["Name"] != "Alice" {
		t.Errorf("Expected Name to be 'Alice', got '%s'", data[1]["Name"])
	}
	if data[1]["Age"] != "25" {
		t.Errorf("Expected Age to be '25', got '%s'", data[1]["Age"])
	}
}

// TestExcel_ReadToMaps tests the ReadToMaps method
func TestExcel_ReadToMaps(t *testing.T) {
	// Create a test Excel file
	excel := New("TestSheet")

	headers := []string{"Name", "Age"}
	data := [][]interface{}{
		{"John", 30},
		{"Alice", 25},
	}

	err := excel.WriteTable(headers, data)
	if err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}

	// Read data
	result, err := excel.ReadToMaps()
	if err != nil {
		t.Fatalf("ReadToMaps returned error: %v", err)
	}

	// Verify data
	if len(result) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(result))
	}

	if result[0]["Name"] != "John" {
		t.Errorf("Expected Name to be 'John', got '%s'", result[0]["Name"])
	}
	if result[0]["Age"] != "30" {
		t.Errorf("Expected Age to be '30', got '%s'", result[0]["Age"])
	}
	if result[1]["Name"] != "Alice" {
		t.Errorf("Expected Name to be 'Alice', got '%s'", result[1]["Name"])
	}
	if result[1]["Age"] != "25" {
		t.Errorf("Expected Age to be '25', got '%s'", result[1]["Age"])
	}
}
