package streamloader_test

import (
	. "github.com/JBrVJxsc/xk6-streamloader"
)
import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWriteMultipleJsonLinesToArrayFile(t *testing.T) {
	loader := StreamLoader{}
	tempDir := t.TempDir() // Create temporary directory for test files

	// Test data
	batch1 := `{"id":1,"name":"Alice"}
{"id":2,"name":"Bob"}`

	batch2 := `{"id":3,"name":"Charlie"}
{"id":4,"name":"Dave","details":{"age":30}}`

	expectedObjects := []interface{}{
		map[string]interface{}{"id": float64(1), "name": "Alice"},
		map[string]interface{}{"id": float64(2), "name": "Bob"},
		map[string]interface{}{"id": float64(3), "name": "Charlie"},
		map[string]interface{}{"id": float64(4), "name": "Dave", "details": map[string]interface{}{"age": float64(30)}},
	}

	tests := []struct {
		name        string
		jsonLines   []string
		bufferSize  []int
		expectCount int
		expectError bool
	}{
		{
			name:        "Empty array",
			jsonLines:   []string{},
			expectCount: 0,
		},
		{
			name:        "Single batch",
			jsonLines:   []string{batch1},
			expectCount: 2,
		},
		{
			name:        "Multiple batches",
			jsonLines:   []string{batch1, batch2},
			expectCount: 4,
		},
		{
			name:        "With empty strings",
			jsonLines:   []string{batch1, "", batch2},
			expectCount: 4,
		},
		{
			name:        "With custom buffer size",
			jsonLines:   []string{batch1, batch2},
			bufferSize:  []int{128}, // 128 bytes
			expectCount: 4,
		},
		{
			name:        "Invalid JSON",
			jsonLines:   []string{batch1, `{invalid json}`},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, "multiple_"+tt.name+".json")

			var count int
			var err error
			if len(tt.bufferSize) > 0 {
				count, err = loader.WriteMultipleJsonLinesToArrayFile(tt.jsonLines, outputPath, tt.bufferSize[0])
			} else {
				count, err = loader.WriteMultipleJsonLinesToArrayFile(tt.jsonLines, outputPath)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("WriteMultipleJsonLinesToArrayFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check count is correct
				if count != tt.expectCount {
					t.Errorf("WriteMultipleJsonLinesToArrayFile() count = %v, want %v", count, tt.expectCount)
				}

				if tt.expectCount > 0 {
					// Read the output file and verify it's a valid JSON array
					content, err := os.ReadFile(outputPath)
					if err != nil {
						t.Fatalf("Failed to read output file: %v", err)
					}

					// Parse the JSON array
					var result []interface{}
					if err := json.Unmarshal(content, &result); err != nil {
						t.Errorf("Output is not a valid JSON array: %v", err)
					}

					// Check the length of the array
					if len(result) != tt.expectCount {
						t.Errorf("JSON array has %d elements, want %d", len(result), tt.expectCount)
					}

					// For the full test case, verify the objects match the expected ones
					if tt.name == "Multiple batches" {
						for i, expectedObj := range expectedObjects {
							if !reflect.DeepEqual(result[i], expectedObj) {
								t.Errorf("Object at index %d doesn't match expected value", i)
							}
						}
					}
				}
			}
		})
	}
}

func TestWriteMultipleCompressedJsonLinesToArrayFile(t *testing.T) {
	loader := StreamLoader{}
	tempDir := t.TempDir() // Create temporary directory for test files

	// Create test objects
	objects1 := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice"},
		map[string]interface{}{"id": 2, "name": "Bob"},
	}

	objects2 := []interface{}{
		map[string]interface{}{"id": 3, "name": "Charlie", "details": map[string]interface{}{"age": 30}},
		map[string]interface{}{"id": 4, "name": "Dave"},
	}

	// Compress the objects to get compressed JSON lines strings
	compressed1, err := loader.ObjectsToCompressedJsonLines(objects1)
	if err != nil {
		t.Fatalf("Failed to compress objects1: %v", err)
	}

	compressed2, err := loader.ObjectsToCompressedJsonLines(objects2)
	if err != nil {
		t.Fatalf("Failed to compress objects2: %v", err)
	}

	// Create invalid compressed data
	invalidCompressed := "not-valid-base64-encoded-gzip-data"

	tests := []struct {
		name             string
		compressedJsonLines []string
		bufferSize       []int
		expectCount      int
		expectError      bool
	}{
		{
			name:             "Empty array",
			compressedJsonLines: []string{},
			expectCount:      0,
		},
		{
			name:             "Single batch",
			compressedJsonLines: []string{compressed1},
			expectCount:      2,
		},
		{
			name:             "Multiple batches",
			compressedJsonLines: []string{compressed1, compressed2},
			expectCount:      4,
		},
		{
			name:             "With empty strings",
			compressedJsonLines: []string{compressed1, "", compressed2},
			expectCount:      4,
		},
		{
			name:             "With custom buffer size",
			compressedJsonLines: []string{compressed1, compressed2},
			bufferSize:       []int{128}, // 128 bytes
			expectCount:      4,
		},
		{
			name:             "Invalid compressed data",
			compressedJsonLines: []string{compressed1, invalidCompressed},
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, "multiple_compressed_"+tt.name+".json")

			var count int
			var err error
			if len(tt.bufferSize) > 0 {
				count, err = loader.WriteMultipleCompressedJsonLinesToArrayFile(tt.compressedJsonLines, outputPath, tt.bufferSize[0])
			} else {
				count, err = loader.WriteMultipleCompressedJsonLinesToArrayFile(tt.compressedJsonLines, outputPath)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("WriteMultipleCompressedJsonLinesToArrayFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check count is correct
				if count != tt.expectCount {
					t.Errorf("WriteMultipleCompressedJsonLinesToArrayFile() count = %v, want %v", count, tt.expectCount)
				}

				if tt.expectCount > 0 {
					// Read the output file and verify it's a valid JSON array
					content, err := os.ReadFile(outputPath)
					if err != nil {
						t.Fatalf("Failed to read output file: %v", err)
					}

					// Parse the JSON array
					var result []interface{}
					if err := json.Unmarshal(content, &result); err != nil {
						t.Errorf("Output is not a valid JSON array: %v", err)
					}

					// Check the length of the array
					if len(result) != tt.expectCount {
						t.Errorf("JSON array has %d elements, want %d", len(result), tt.expectCount)
					}

					// For the multiple batches test, verify all objects were correctly decompressed and written
					if tt.name == "Multiple batches" {
						allObjects := append(objects1, objects2...)
						for i, obj := range allObjects {
							originalJSON, _ := json.Marshal(obj)
							resultJSON, _ := json.Marshal(result[i])

							var originalMap, resultMap map[string]interface{}
							json.Unmarshal(originalJSON, &originalMap)
							json.Unmarshal(resultJSON, &resultMap)

							if !reflect.DeepEqual(originalMap, resultMap) {
								t.Errorf("Object at index %d doesn't match after decompression", i)
							}
						}
					}
				}
			}
		})
	}
}

func TestMultipleCompressedJsonLinesRoundtrip(t *testing.T) {
	loader := StreamLoader{}
	tempDir := t.TempDir() // Create temporary directory for test files
	
	// Create test objects with different structures and content types
	testCases := []struct {
		name    string
		objects []interface{}
	}{
		{
			name:    "Simple objects",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
				map[string]interface{}{"id": 2, "name": "Bob"},
			},
		},
		{
			name:    "Nested objects",
			objects: []interface{}{
				map[string]interface{}{
					"id":   3, 
					"name": "Charlie",
					"details": map[string]interface{}{
						"age":  30,
						"city": "New York",
					},
				},
			},
		},
		{
			name:    "Objects with arrays",
			objects: []interface{}{
				map[string]interface{}{
					"id":     4,
					"name":   "Dave",
					"skills": []string{"Java", "Python", "Go"},
				},
			},
		},
		{
			name:    "Special characters",
			objects: []interface{}{
				map[string]interface{}{
					"id":          5,
					"description": "Product with \"quotes\" and commas, plus other chars: !@#$%^&*()",
					"html":        "<div>Some HTML content</div>",
				},
			},
		},
	}
	
	// Compress each batch of objects
	compressedBatches := make([]string, len(testCases))
	var allObjects []interface{}
	
	for i, tc := range testCases {
		compressed, err := loader.ObjectsToCompressedJsonLines(tc.objects)
		if err != nil {
			t.Fatalf("Failed to compress objects for %s: %v", tc.name, err)
		}
		compressedBatches[i] = compressed
		allObjects = append(allObjects, tc.objects...)
	}
	
	// Write all batches to a single file
	outputPath := filepath.Join(tempDir, "round_trip_test.json")
	count, err := loader.WriteMultipleCompressedJsonLinesToArrayFile(compressedBatches, outputPath)
	
	if err != nil {
		t.Fatalf("WriteMultipleCompressedJsonLinesToArrayFile failed: %v", err)
	}
	
	// Verify the count matches the total number of objects
	expectedCount := 0
	for _, tc := range testCases {
		expectedCount += len(tc.objects)
	}
	
	if count != expectedCount {
		t.Errorf("Expected %d objects, got %d", expectedCount, count)
	}
	
	// Read the output file and verify its contents
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	var result []interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatalf("Output is not a valid JSON array: %v", err)
	}
	
	// Check the length of the array
	if len(result) != expectedCount {
		t.Errorf("JSON array has %d elements, want %d", len(result), expectedCount)
	}
	
	// Compare original objects with the deserialized ones
	for i, originalObj := range allObjects {
		originalJSON, _ := json.Marshal(originalObj)
		resultJSON, _ := json.Marshal(result[i])
		
		var originalMap, resultMap map[string]interface{}
		json.Unmarshal(originalJSON, &originalMap)
		json.Unmarshal(resultJSON, &resultMap)
		
		if !reflect.DeepEqual(originalMap, resultMap) {
			t.Errorf("Object at index %d doesn't match after roundtrip", i)
			t.Errorf("Original: %v", originalMap)
			t.Errorf("Result: %v", resultMap)
		}
	}
}