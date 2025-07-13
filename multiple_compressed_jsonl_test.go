package streamloader_test

import (
	. "github.com/JBrVJxsc/xk6-streamloader"
)
import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Helper function to create temporary files for testing
func createTempFile(t *testing.T) string {
	file, err := os.CreateTemp("", "streamloader_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	file.Close()
	return file.Name()
}

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

func TestMultipleCompressedJsonLinesToObjects(t *testing.T) {
	loader := StreamLoader{}

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

	tests := []struct {
		name              string
		compressedStrings []string
		expectCount       int
		expectError       bool
	}{
		{
			name:              "Empty array",
			compressedStrings: []string{},
			expectCount:       0,
		},
		{
			name:              "Single batch",
			compressedStrings: []string{compressed1},
			expectCount:       2,
		},
		{
			name:              "Multiple batches",
			compressedStrings: []string{compressed1, compressed2},
			expectCount:       4,
		},
		{
			name:              "With empty strings",
			compressedStrings: []string{compressed1, "", compressed2},
			expectCount:       4,
		},
		{
			name:              "Invalid base64",
			compressedStrings: []string{compressed1, "invalid_base64!!!"},
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := loader.MultipleCompressedJsonLinesToObjects(tt.compressedStrings)

			if (err != nil) != tt.expectError {
				t.Errorf("MultipleCompressedJsonLinesToObjects() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check count is correct
				if len(result) != tt.expectCount {
					t.Errorf("MultipleCompressedJsonLinesToObjects() returned %d objects, want %d", len(result), tt.expectCount)
				}

				// For the multiple batches test, verify all objects were correctly decompressed
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
							t.Errorf("Original: %v", originalMap)
							t.Errorf("Result: %v", resultMap)
						}
					}
				}

				// For single batch test, verify the objects match
				if tt.name == "Single batch" {
					for i, obj := range objects1 {
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
		})
	}

	// Test roundtrip with complex data
	t.Run("Complex data roundtrip", func(t *testing.T) {
		complexObjects := []interface{}{
			map[string]interface{}{
				"id":   5,
				"name": "Eve",
				"profile": map[string]interface{}{
					"age": 28,
					"address": map[string]interface{}{
						"street": "123 Main St",
						"city":   "New York",
					},
					"hobbies": []string{"reading", "cycling"},
				},
			},
			map[string]interface{}{
				"id":          6,
				"description": "Product with \"quotes\" and commas, plus other chars: !@#$%^&*()",
				"active":      true,
				"score":       95.7,
			},
		}

		// Split into two batches
		batch1 := complexObjects[:1]
		batch2 := complexObjects[1:]

		compressedBatch1, err := loader.ObjectsToCompressedJsonLines(batch1)
		if err != nil {
			t.Fatalf("Failed to compress batch1: %v", err)
		}

		compressedBatch2, err := loader.ObjectsToCompressedJsonLines(batch2)
		if err != nil {
			t.Fatalf("Failed to compress batch2: %v", err)
		}

		// Test the new function
		result, err := loader.MultipleCompressedJsonLinesToObjects([]string{compressedBatch1, compressedBatch2})
		if err != nil {
			t.Fatalf("MultipleCompressedJsonLinesToObjects failed: %v", err)
		}

		if len(result) != len(complexObjects) {
			t.Errorf("Expected %d objects, got %d", len(complexObjects), len(result))
		}

		// Verify each object matches
		for i, originalObj := range complexObjects {
			originalJSON, _ := json.Marshal(originalObj)
			resultJSON, _ := json.Marshal(result[i])

			var originalMap, resultMap map[string]interface{}
			json.Unmarshal(originalJSON, &originalMap)
			json.Unmarshal(resultJSON, &resultMap)

			if !reflect.DeepEqual(originalMap, resultMap) {
				t.Errorf("Complex object at index %d doesn't match after decompression", i)
				t.Errorf("Original: %v", originalMap)
				t.Errorf("Result: %v", resultMap)
			}
		}
	})
}

func TestWriteWeightedMultipleCompressedJsonLinesToArrayFile(t *testing.T) {
	loader := StreamLoader{}

	// Create test objects for different scenarios
	batch1Objects := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice"},
		map[string]interface{}{"id": 2, "name": "Bob"},
	}

	batch2Objects := []interface{}{
		map[string]interface{}{"id": 3, "name": "Charlie"},
		map[string]interface{}{"id": 4, "name": "Dave"},
		map[string]interface{}{"id": 5, "name": "Eve"},
		map[string]interface{}{"id": 6, "name": "Frank"},
		map[string]interface{}{"id": 7, "name": "Grace"},
	}

	batch3Objects := []interface{}{
		map[string]interface{}{"id": 8, "name": "Henry"},
	}

	// Compress the test batches
	compressedBatch1, err := loader.ObjectsToCompressedJsonLines(batch1Objects)
	if err != nil {
		t.Fatalf("Failed to compress batch1: %v", err)
	}

	compressedBatch2, err := loader.ObjectsToCompressedJsonLines(batch2Objects)
	if err != nil {
		t.Fatalf("Failed to compress batch2: %v", err)
	}

	compressedBatch3, err := loader.ObjectsToCompressedJsonLines(batch3Objects)
	if err != nil {
		t.Fatalf("Failed to compress batch3: %v", err)
	}

	t.Run("Empty array", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		count, err := loader.WriteWeightedMultipleCompressedJsonLinesToArrayFile([][]interface{}{}, tempFile)
		if err != nil {
			t.Errorf("Expected no error for empty array, got: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count 0 for empty array, got: %d", count)
		}

		// Verify file contains empty JSON array
		content, _ := os.ReadFile(tempFile)
		if string(content) != "[]" {
			t.Errorf("Expected empty JSON array, got: %s", string(content))
		}
	})

	t.Run("Equal weight (count == weight)", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		// batch1 has 2 objects, weight 2 -> keep all 2 objects
		weightedBatches := [][]interface{}{
			{[]string{compressedBatch1}, 2},
		}

		count, err := loader.WriteWeightedMultipleCompressedJsonLinesToArrayFile(weightedBatches, tempFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got: %d", count)
		}

		// Verify output
		var result []map[string]interface{}
		content, _ := os.ReadFile(tempFile)
		json.Unmarshal(content, &result)

		if len(result) != 2 {
			t.Errorf("Expected 2 objects in output, got: %d", len(result))
		}
		if result[0]["name"] != "Alice" || result[1]["name"] != "Bob" {
			t.Error("Objects not preserved correctly")
		}
	})

	t.Run("Oversample (count > weight)", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		// batch2 has 5 objects, weight 3 -> slice to keep first 3 objects
		weightedBatches := [][]interface{}{
			{[]string{compressedBatch2}, 3},
		}

		count, err := loader.WriteWeightedMultipleCompressedJsonLinesToArrayFile(weightedBatches, tempFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count 3, got: %d", count)
		}

		// Verify output contains first 3 objects
		var result []map[string]interface{}
		content, _ := os.ReadFile(tempFile)
		json.Unmarshal(content, &result)

		if len(result) != 3 {
			t.Errorf("Expected 3 objects in output, got: %d", len(result))
		}
		expectedNames := []string{"Charlie", "Dave", "Eve"}
		for i, expected := range expectedNames {
			if result[i]["name"] != expected {
				t.Errorf("Expected name %s at index %d, got: %s", expected, i, result[i]["name"])
			}
		}
	})

	t.Run("Undersample (count < weight)", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		// batch3 has 1 object, weight 4 -> duplicate cyclically: [Henry, Henry, Henry, Henry]
		weightedBatches := [][]interface{}{
			{[]string{compressedBatch3}, 4},
		}

		count, err := loader.WriteWeightedMultipleCompressedJsonLinesToArrayFile(weightedBatches, tempFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count != 4 {
			t.Errorf("Expected count 4, got: %d", count)
		}

		// Verify output contains 4 duplicated objects
		var result []map[string]interface{}
		content, _ := os.ReadFile(tempFile)
		json.Unmarshal(content, &result)

		if len(result) != 4 {
			t.Errorf("Expected 4 objects in output, got: %d", len(result))
		}
		for i := 0; i < 4; i++ {
			if result[i]["name"] != "Henry" || result[i]["id"].(float64) != 8 {
				t.Errorf("Expected duplicated Henry object at index %d, got: %v", i, result[i])
			}
		}
	})

	t.Run("Complex duplication pattern", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		// batch1 has 2 objects [Alice, Bob], weight 5 -> [Alice, Bob, Alice, Bob, Alice]
		weightedBatches := [][]interface{}{
			{[]string{compressedBatch1}, 5},
		}

		count, err := loader.WriteWeightedMultipleCompressedJsonLinesToArrayFile(weightedBatches, tempFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count != 5 {
			t.Errorf("Expected count 5, got: %d", count)
		}

		// Verify cyclic duplication pattern
		var result []map[string]interface{}
		content, _ := os.ReadFile(tempFile)
		json.Unmarshal(content, &result)

		if len(result) != 5 {
			t.Errorf("Expected 5 objects in output, got: %d", len(result))
		}
		expectedPattern := []string{"Alice", "Bob", "Alice", "Bob", "Alice"}
		for i, expected := range expectedPattern {
			if result[i]["name"] != expected {
				t.Errorf("Expected name %s at index %d, got: %s", expected, i, result[i]["name"])
			}
		}
	})

	t.Run("Multiple weighted batches", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		// Combine multiple batches with different weights
		weightedBatches := [][]interface{}{
			{[]string{compressedBatch1}, 1}, // 2 objects -> 1 object (Alice)
			{[]string{compressedBatch2}, 2}, // 5 objects -> 2 objects (Charlie, Dave)
			{[]string{compressedBatch3}, 3}, // 1 object -> 3 objects (Henry, Henry, Henry)
		}

		count, err := loader.WriteWeightedMultipleCompressedJsonLinesToArrayFile(weightedBatches, tempFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count != 6 {
			t.Errorf("Expected count 6, got: %d", count)
		}

		// Verify combined output
		var result []map[string]interface{}
		content, _ := os.ReadFile(tempFile)
		json.Unmarshal(content, &result)

		if len(result) != 6 {
			t.Errorf("Expected 6 objects in output, got: %d", len(result))
		}

		// Check specific pattern: Alice, Charlie, Dave, Henry, Henry, Henry
		expectedNames := []string{"Alice", "Charlie", "Dave", "Henry", "Henry", "Henry"}
		for i, expected := range expectedNames {
			if result[i]["name"] != expected {
				t.Errorf("Expected name %s at index %d, got: %s", expected, i, result[i]["name"])
			}
		}
	})

	t.Run("Zero and negative weights", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		// Test with zero and negative weights (should be skipped)
		weightedBatches := [][]interface{}{
			{[]string{compressedBatch1}, 0},  // Should be skipped
			{[]string{compressedBatch2}, -1}, // Should be skipped
			{[]string{compressedBatch3}, 2},  // Should produce 2 Henry objects
		}

		count, err := loader.WriteWeightedMultipleCompressedJsonLinesToArrayFile(weightedBatches, tempFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got: %d", count)
		}

		// Verify only valid weighted batch was processed
		var result []map[string]interface{}
		content, _ := os.ReadFile(tempFile)
		json.Unmarshal(content, &result)

		if len(result) != 2 {
			t.Errorf("Expected 2 objects in output, got: %d", len(result))
		}
		for i := 0; i < 2; i++ {
			if result[i]["name"] != "Henry" {
				t.Errorf("Expected Henry at index %d, got: %s", i, result[i]["name"])
			}
		}
	})

	t.Run("Invalid input format", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		// Test with invalid entry format
		weightedBatches := [][]interface{}{
			{[]string{compressedBatch1}}, // Missing weight
		}

		_, err := loader.WriteWeightedMultipleCompressedJsonLinesToArrayFile(weightedBatches, tempFile)
		if err == nil {
			t.Error("Expected error for invalid entry format")
		}
		if !strings.Contains(err.Error(), "expected [multipleCompressedJsonLines, weight]") {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Invalid weight type", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		// Test with invalid weight type
		weightedBatches := [][]interface{}{
			{[]string{compressedBatch1}, "invalid"}, // String instead of number
		}

		_, err := loader.WriteWeightedMultipleCompressedJsonLinesToArrayFile(weightedBatches, tempFile)
		if err == nil {
			t.Error("Expected error for invalid weight type")
		}
		if !strings.Contains(err.Error(), "expected number") {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Invalid base64 data", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		// Test with invalid base64 data
		weightedBatches := [][]interface{}{
			{[]string{"invalid-base64-data!!!"}, 2},
		}

		_, err := loader.WriteWeightedMultipleCompressedJsonLinesToArrayFile(weightedBatches, tempFile)
		if err == nil {
			t.Error("Expected error for invalid base64 data")
		}
		if !strings.Contains(err.Error(), "failed to decode base64 data") {
			t.Errorf("Expected base64 decode error, got: %v", err)
		}
	})

	t.Run("Float weight compatibility", func(t *testing.T) {
		tempFile := createTempFile(t)
		defer os.Remove(tempFile)

		// Test with float weight (common from JavaScript)
		weightedBatches := [][]interface{}{
			{[]string{compressedBatch1}, 3.0}, // Float weight should be converted to int
		}

		count, err := loader.WriteWeightedMultipleCompressedJsonLinesToArrayFile(weightedBatches, tempFile)
		if err != nil {
			t.Errorf("Unexpected error with float weight: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count 3, got: %d", count)
		}
	})
}