package streamloader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrimLeadingSpaceOption(t *testing.T) {
	// Create a temporary CSV file with leading spaces
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "trim_test.csv")

	// CSV with leading spaces in fields
	csvContent := `id,name,value
1, Product with space,100
2,  Double space product,200
3,   Triple space,300`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	loader := StreamLoader{}

	// Test with TrimLeadingSpace=true (default)
	t.Run("With TrimLeadingSpace=true", func(t *testing.T) {
		records, err := loader.LoadCSV(csvPath, CsvOptions{
			TrimLeadingSpace: true,
			LazyQuotes:       true,
			ReuseRecord:      true,
		})
		if err != nil {
			t.Fatalf("LoadCSV failed: %v", err)
		}

		// Check that leading spaces were trimmed
		if records[1][1] != "Product with space" {
			t.Errorf("Expected trimmed value 'Product with space', got '%s'", records[1][1])
		}
		if records[2][1] != "Double space product" {
			t.Errorf("Expected trimmed value 'Double space product', got '%s'", records[2][1])
		}
		if records[3][1] != "Triple space" {
			t.Errorf("Expected trimmed value 'Triple space', got '%s'", records[3][1])
		}
	})

	// Test with TrimLeadingSpace=false
	t.Run("With TrimLeadingSpace=false", func(t *testing.T) {
		records, err := loader.LoadCSV(csvPath, CsvOptions{
			TrimLeadingSpace: false,
			LazyQuotes:       true,
			ReuseRecord:      true,
		})
		if err != nil {
			t.Fatalf("LoadCSV failed: %v", err)
		}

		// Check that leading spaces were preserved
		if records[1][1] != " Product with space" {
			t.Errorf("Expected untrimmed value ' Product with space', got '%s'", records[1][1])
		}
		if records[2][1] != "  Double space product" {
			t.Errorf("Expected untrimmed value '  Double space product', got '%s'", records[2][1])
		}
		if records[3][1] != "   Triple space" {
			t.Errorf("Expected untrimmed value '   Triple space', got '%s'", records[3][1])
		}
	})
}

func TestReuseRecordOption(t *testing.T) {
	// Create a temporary CSV file
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "reuse_test.csv")

	// Simple CSV to test with
	csvContent := `id,name,value
1,Product 1,100
2,Product 2,200
3,Product 3,300`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	loader := StreamLoader{}

	// Test with ReuseRecord=true (this mainly tests that it doesn't crash)
	t.Run("With ReuseRecord=true", func(t *testing.T) {
		records, err := loader.LoadCSV(csvPath, CsvOptions{
			TrimLeadingSpace: true,
			LazyQuotes:       true,
			ReuseRecord:      true,
		})
		if err != nil {
			t.Fatalf("LoadCSV failed: %v", err)
		}

		// Just verify we got the right number of records
		if len(records) != 4 { // 1 header + 3 data rows
			t.Errorf("Expected 4 records, got %d", len(records))
		}
	})

	// Test with ReuseRecord=false
	t.Run("With ReuseRecord=false", func(t *testing.T) {
		records, err := loader.LoadCSV(csvPath, CsvOptions{
			TrimLeadingSpace: true,
			LazyQuotes:       true,
			ReuseRecord:      false,
		})
		if err != nil {
			t.Fatalf("LoadCSV failed: %v", err)
		}

		// Just verify we got the right number of records
		if len(records) != 4 { // 1 header + 3 data rows
			t.Errorf("Expected 4 records, got %d", len(records))
		}
	})
}

func TestOptionsCombo(t *testing.T) {
	// Test all options in combination
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "combo_test.csv")

	// CSV with various challenges
	csvContent := `id,name,value
1, "Product, with comma",100
2,  Product with "quotes",200
3,"   Quoted with spaces   ",300`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	loader := StreamLoader{}
	
	options := CsvOptions{
		LazyQuotes:       true,
		TrimLeadingSpace: true,
		ReuseRecord:      true,
	}
	
	records, err := loader.LoadCSV(csvPath, options)
	if err != nil {
		t.Fatalf("LoadCSV failed with all options: %v", err)
	}
	
	if len(records) != 4 {
		t.Errorf("Expected 4 records, got %d", len(records))
	}
}