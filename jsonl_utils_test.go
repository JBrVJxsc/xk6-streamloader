package streamloader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestObjectsToJsonLines(t *testing.T) {
	loader := StreamLoader{}

	tests := []struct {
		name        string
		objects     []interface{}
		expectError bool
	}{
		{
			name:    "Empty array",
			objects: []interface{}{},
		},
		{
			name: "Single object",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
			},
		},
		{
			name: "Multiple objects",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
				map[string]interface{}{"id": 2, "name": "Bob"},
				map[string]interface{}{"id": 3, "name": "Charlie"},
			},
		},
		{
			name: "Nested objects",
			objects: []interface{}{
				map[string]interface{}{
					"id":   1,
					"name": "Alice",
					"details": map[string]interface{}{
						"age":  30,
						"city": "New York",
					},
				},
				map[string]interface{}{
					"id":   2,
					"name": "Bob",
					"details": map[string]interface{}{
						"age":  25,
						"city": "Los Angeles",
					},
				},
			},
		},
		{
			name: "Object with array",
			objects: []interface{}{
				map[string]interface{}{
					"id":     1,
					"name":   "Alice",
					"skills": []string{"Java", "Python", "Go"},
				},
			},
		},
		{
			name: "Object with special characters",
			objects: []interface{}{
				map[string]interface{}{
					"id":          1,
					"description": "This has \"quotes\", commas, and other stuff: !@#$%^&*()",
					"html":        "<div>Some HTML content</div>",
				},
			},
		},
		{
			name: "Object with non-ASCII characters",
			objects: []interface{}{
				map[string]interface{}{
					"id":       1,
					"name":     "José Müller",
					"location": "北京市",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loader.ObjectsToJsonLines(tt.objects)
			if (err != nil) != tt.expectError {
				t.Errorf("ObjectsToJsonLines() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Verify that each line can be parsed back into a valid JSON object
			if got != "" {
				lines := parseJsonLines(t, got)
				if len(lines) != len(tt.objects) {
					t.Errorf("Expected %d JSON objects, parsed %d", len(tt.objects), len(lines))
				}
				
				// Verify that each parsed object contains the same data as the original
				for i, obj := range lines {
					// Marshal both original and parsed object to normalize the JSON
					originalJSON, _ := json.Marshal(tt.objects[i])
					parsedJSON, _ := json.Marshal(obj)
					
					// Unmarshal into maps for comparison
					var originalMap, parsedMap map[string]interface{}
					json.Unmarshal(originalJSON, &originalMap)
					json.Unmarshal(parsedJSON, &parsedMap)
					
					// Compare the maps
					if !reflect.DeepEqual(originalMap, parsedMap) {
						t.Errorf("Object at index %d doesn't match after serialization/parsing", i)
					}
				}
			}
		})
	}
}

func TestWriteJsonLinesToArrayFile(t *testing.T) {
	loader := StreamLoader{}
	tempDir := t.TempDir() // Create temporary directory for test files

	tests := []struct {
		name        string
		jsonLines   string
		bufferSize  []int
		expectCount int
		expectError bool
	}{
		{
			name:        "Empty JSON lines",
			jsonLines:   "",
			expectCount: 0,
		},
		{
			name:        "Single JSON object",
			jsonLines:   `{"id":1,"name":"Alice"}`,
			expectCount: 1,
		},
		{
			name:        "Multiple JSON objects",
			jsonLines:   "{\"id\":1,\"name\":\"Alice\"}\n{\"id\":2,\"name\":\"Bob\"}\n{\"id\":3,\"name\":\"Charlie\"}",
			expectCount: 3,
		},
		{
			name:        "JSON objects with blank lines",
			jsonLines:   "{\"id\":1,\"name\":\"Alice\"}\n\n{\"id\":2,\"name\":\"Bob\"}\n\n",
			expectCount: 2,
		},
		{
			name:        "JSON objects with whitespace",
			jsonLines:   "  {\"id\":1,\"name\":\"Alice\"}  \n  {\"id\":2,\"name\":\"Bob\"}  ",
			expectCount: 2,
		},
		{
			name:        "Complex JSON objects",
			jsonLines:   "{\"id\":1,\"name\":\"Alice\",\"details\":{\"age\":30,\"city\":\"New York\"}}\n{\"id\":2,\"name\":\"Bob\",\"skills\":[\"Java\",\"Python\"]}",
			expectCount: 2,
		},
		{
			name:        "With custom buffer size",
			jsonLines:   "{\"id\":1,\"name\":\"Alice\"}\n{\"id\":2,\"name\":\"Bob\"}",
			bufferSize:  []int{128}, // 128 bytes
			expectCount: 2,
		},
		{
			name:        "Invalid JSON",
			jsonLines:   `{"id":1,"name":"Alice"}\n{this-is-not-json}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, fmt.Sprintf("test_%s.json", tt.name))

			var count int
			var err error
			if len(tt.bufferSize) > 0 {
				count, err = loader.WriteJsonLinesToArrayFile(tt.jsonLines, outputPath, tt.bufferSize[0])
			} else {
				count, err = loader.WriteJsonLinesToArrayFile(tt.jsonLines, outputPath)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("WriteJsonLinesToArrayFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check count is correct
				if count != tt.expectCount {
					t.Errorf("WriteJsonLinesToArrayFile() count = %v, want %v", count, tt.expectCount)
				}

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
			}
		})
	}
}

func TestCombineJsonArrayFiles(t *testing.T) {
	loader := StreamLoader{}
	tempDir := t.TempDir() // Create temporary directory for test files

	// Create test files
	testFiles := []struct {
		name    string
		content string
	}{
		{"empty.json", "[]"},
		{"single.json", "[{\"id\":1,\"name\":\"Alice\"}]"},
		{"multiple.json", "[{\"id\":2,\"name\":\"Bob\"},{\"id\":3,\"name\":\"Charlie\"}]"},
		{"complex.json", "[{\"id\":4,\"details\":{\"age\":30}},{\"id\":5,\"skills\":[\"Java\",\"Go\"]}]"},
		{"invalid.json", "{not-valid-json}"},
		{"not_array.json", "{\"key\":\"value\"}"},
	}

	// Create test files
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file.name)
		if err := os.WriteFile(filePath, []byte(file.content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file.name, err)
		}
	}

	tests := []struct {
		name          string
		inputPaths    []string
		expectCount   int
		expectError   bool
		bufferSize    []int
		expectedItems []map[string]interface{}
	}{
		{
			name:        "Empty file",
			inputPaths:  []string{filepath.Join(tempDir, "empty.json")},
			expectCount: 0,
		},
		{
			name:        "Single file with one object",
			inputPaths:  []string{filepath.Join(tempDir, "single.json")},
			expectCount: 1,
			expectedItems: []map[string]interface{}{
				{"id": float64(1), "name": "Alice"},
			},
		},
		{
			name:        "Single file with multiple objects",
			inputPaths:  []string{filepath.Join(tempDir, "multiple.json")},
			expectCount: 2,
			expectedItems: []map[string]interface{}{
				{"id": float64(2), "name": "Bob"},
				{"id": float64(3), "name": "Charlie"},
			},
		},
		{
			name:        "Multiple files",
			inputPaths:  []string{filepath.Join(tempDir, "single.json"), filepath.Join(tempDir, "multiple.json")},
			expectCount: 3,
			expectedItems: []map[string]interface{}{
				{"id": float64(1), "name": "Alice"},
				{"id": float64(2), "name": "Bob"},
				{"id": float64(3), "name": "Charlie"},
			},
		},
		{
			name:        "Complex objects",
			inputPaths:  []string{filepath.Join(tempDir, "complex.json")},
			expectCount: 2,
		},
		{
			name:        "Custom buffer size",
			inputPaths:  []string{filepath.Join(tempDir, "single.json")},
			bufferSize:  []int{128}, // 128 bytes
			expectCount: 1,
		},
		{
			name:        "Invalid file path",
			inputPaths:  []string{filepath.Join(tempDir, "nonexistent.json")},
			expectError: true,
		},
		{
			name:        "Invalid JSON file",
			inputPaths:  []string{filepath.Join(tempDir, "invalid.json")},
			expectError: true,
		},
		{
			name:        "Not a JSON array",
			inputPaths:  []string{filepath.Join(tempDir, "not_array.json")},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, fmt.Sprintf("combined_%s.json", tt.name))

			var count int
			var err error
			if len(tt.bufferSize) > 0 {
				count, err = loader.CombineJsonArrayFiles(tt.inputPaths, outputPath, tt.bufferSize[0])
			} else {
				count, err = loader.CombineJsonArrayFiles(tt.inputPaths, outputPath)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("CombineJsonArrayFiles() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check count is correct
				if count != tt.expectCount {
					t.Errorf("CombineJsonArrayFiles() count = %v, want %v", count, tt.expectCount)
				}

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

				// Check specific items if expected
				if len(tt.expectedItems) > 0 {
					for i, expectedItem := range tt.expectedItems {
						if i >= len(result) {
							t.Errorf("Missing expected item at index %d", i)
							continue
						}

						actualItem, ok := result[i].(map[string]interface{})
						if !ok {
							t.Errorf("Item at index %d is not a map", i)
							continue
						}

						for key, expectedValue := range expectedItem {
							actualValue, exists := actualItem[key]
							if !exists {
								t.Errorf("Item at index %d missing key %s", i, key)
								continue
							}
							if !reflect.DeepEqual(actualValue, expectedValue) {
								t.Errorf("Item at index %d, key %s: got %v, want %v", i, key, actualValue, expectedValue)
							}
						}
					}
				}
			}
		})
	}
}

func TestWriteObjectsToJsonArrayFile(t *testing.T) {
	loader := StreamLoader{}
	tempDir := t.TempDir() // Create temporary directory for test files

	tests := []struct {
		name        string
		objects     []interface{}
		bufferSize  []int
		expectCount int
		expectError bool
	}{
		{
			name:        "Empty array",
			objects:     []interface{}{},
			expectCount: 0,
		},
		{
			name: "Single object",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
			},
			expectCount: 1,
		},
		{
			name: "Multiple objects",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
				map[string]interface{}{"id": 2, "name": "Bob"},
				map[string]interface{}{"id": 3, "name": "Charlie"},
			},
			expectCount: 3,
		},
		{
			name: "Nested objects",
			objects: []interface{}{
				map[string]interface{}{
					"id": 1,
					"details": map[string]interface{}{
						"age":  30,
						"city": "New York",
					},
				},
			},
			expectCount: 1,
		},
		{
			name: "Object with arrays",
			objects: []interface{}{
				map[string]interface{}{
					"id":     1,
					"skills": []string{"Java", "Python", "Go"},
				},
			},
			expectCount: 1,
		},
		{
			name: "With custom buffer size",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
			},
			bufferSize:  []int{128}, // 128 bytes
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, fmt.Sprintf("objects_%s.json", tt.name))

			var count int
			var err error
			if len(tt.bufferSize) > 0 {
				count, err = loader.WriteObjectsToJsonArrayFile(tt.objects, outputPath, tt.bufferSize[0])
			} else {
				count, err = loader.WriteObjectsToJsonArrayFile(tt.objects, outputPath)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("WriteObjectsToJsonArrayFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check count is correct
				if count != tt.expectCount {
					t.Errorf("WriteObjectsToJsonArrayFile() count = %v, want %v", count, tt.expectCount)
				}

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

				// Compare original objects with the read ones
				if tt.expectCount > 0 {
					// Need to convert both to JSON and back to normalize the representation
					originalJson, _ := json.Marshal(tt.objects)
					resultJson, _ := json.Marshal(result)

					var normalizedOriginal, normalizedResult []interface{}
					json.Unmarshal(originalJson, &normalizedOriginal)
					json.Unmarshal(resultJson, &normalizedResult)

					if !reflect.DeepEqual(normalizedOriginal, normalizedResult) {
						t.Errorf("Objects don't match after roundtrip.\nOriginal: %v\nResult: %v", normalizedOriginal, normalizedResult)
					}
				}
			}
		})
	}
}

func TestRoundtripFunctionality(t *testing.T) {
	loader := StreamLoader{}
	tempDir := t.TempDir() // Create temporary directory for test files

	// Test cases for roundtrip testing
	testCases := []struct {
		name    string
		objects []interface{}
	}{
		{
			name:    "Empty array",
			objects: []interface{}{},
		},
		{
			name: "Simple objects",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
				map[string]interface{}{"id": 2, "name": "Bob"},
			},
		},
		{
			name: "Complex nested objects",
			objects: []interface{}{
				map[string]interface{}{
					"id": 1,
					"user": map[string]interface{}{
						"name":  "Alice",
						"email": "alice@example.com",
						"address": map[string]interface{}{
							"street": "123 Main St",
							"city":   "New York",
							"zip":    "10001",
						},
						"phones": []string{"555-1234", "555-5678"},
					},
					"orders": []interface{}{
						map[string]interface{}{
							"id":    "ORD-001",
							"total": 129.99,
							"items": []interface{}{
								map[string]interface{}{"product": "Laptop", "price": 999.99},
								map[string]interface{}{"product": "Mouse", "price": 29.99},
							},
						},
					},
					"active": true,
					"score":  95.5,
				},
			},
		},
		{
			name: "Objects with special characters",
			objects: []interface{}{
				map[string]interface{}{
					"id":          1,
					"description": "Product with \"quotes\" and commas, plus other chars: !@#$%^&*()",
					"html":        "<div>Some HTML content with <br/> tags</div>",
					"json":        "{\"nested\":\"json string\"}",
				},
			},
		},
		{
			name: "Objects with non-ASCII characters",
			objects: []interface{}{
				map[string]interface{}{
					"id":       1,
					"name":     "José María Müller",
					"location": "北京市朝阳区",
					"emoji":    "😊👍👨‍👩‍👧‍👦", // Emoji and complex Unicode
				},
			},
		},
		{
			name: "Objects with null values",
			objects: []interface{}{
				map[string]interface{}{
					"id":          1,
					"name":        "Alice",
					"description": nil,
					"metadata":    nil,
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Convert objects to JSON lines
			jsonLines, err := loader.ObjectsToJsonLines(tt.objects)
			if err != nil {
				t.Fatalf("ObjectsToJsonLines failed: %v", err)
			}

			// Step 2: Write JSON lines to file
			jsonlFilePath := filepath.Join(tempDir, fmt.Sprintf("roundtrip_%s_lines.jsonl", tt.name))
			if err := os.WriteFile(jsonlFilePath, []byte(jsonLines), 0644); err != nil {
				t.Fatalf("Failed to write JSONL file: %v", err)
			}

			// Step 3: Read JSONL file and convert to JSON array
			jsonlContent, err := os.ReadFile(jsonlFilePath)
			if err != nil {
				t.Fatalf("Failed to read JSONL file: %v", err)
			}

			// Step 4: Write JSON lines to JSON array file
			jsonArrayFilePath := filepath.Join(tempDir, fmt.Sprintf("roundtrip_%s_array.json", tt.name))
			count, err := loader.WriteJsonLinesToArrayFile(string(jsonlContent), jsonArrayFilePath)
			if err != nil {
				t.Fatalf("WriteJsonLinesToArrayFile failed: %v", err)
			}

			// Verify count matches
			if count != len(tt.objects) {
				t.Errorf("Expected %d objects, got %d", len(tt.objects), count)
			}

			// Step 5: Read back the JSON array and compare with original objects
			jsonArrayContent, err := os.ReadFile(jsonArrayFilePath)
			if err != nil {
				t.Fatalf("Failed to read JSON array file: %v", err)
			}

			var result []interface{}
			if err := json.Unmarshal(jsonArrayContent, &result); err != nil {
				t.Fatalf("Failed to unmarshal JSON array: %v", err)
			}

			// Compare original and round-tripped objects
			if len(result) != len(tt.objects) {
				t.Errorf("Object count mismatch: got %d, want %d", len(result), len(tt.objects))
			}

			// Need to normalize both object representations by marshaling to JSON and back
			originalJson, _ := json.Marshal(tt.objects)
			resultJson, _ := json.Marshal(result)

			var normalizedOriginal, normalizedResult []interface{}
			json.Unmarshal(originalJson, &normalizedOriginal)
			json.Unmarshal(resultJson, &normalizedResult)

			if !reflect.DeepEqual(normalizedOriginal, normalizedResult) {
				t.Errorf("Objects don't match after roundtrip.\nOriginal: %v\nResult: %v", normalizedOriginal, normalizedResult)
			}
		})
	}
}

// Helper function to parse a string of JSON lines into an array of objects
func parseJsonLines(t *testing.T, jsonLines string) []interface{} {
	if jsonLines == "" {
		return []interface{}{}
	}

	var result []interface{}
	for _, line := range splitLines(jsonLines) {
		if line == "" {
			continue
		}
		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Fatalf("Failed to parse JSON line: %v\nLine: %s", err, line)
		}
		result = append(result, obj)
	}
	return result
}

// Helper function to split a string into lines
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}
