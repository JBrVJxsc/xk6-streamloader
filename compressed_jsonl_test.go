package streamloader

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestObjectsToCompressedJsonLines(t *testing.T) {
	loader := StreamLoader{}

	tests := []struct {
		name             string
		objects          []interface{}
		compressionLevel []int
		expectError      bool
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
			name: "With compression level 0 (no compression)",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
			},
			compressionLevel: []int{gzip.NoCompression},
		},
		{
			name: "With compression level 9 (best compression)",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
			},
			compressionLevel: []int{gzip.BestCompression},
		},
		{
			name: "Invalid compression level (will use default)",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
			},
			compressionLevel: []int{10}, // Invalid level, should use default
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
					"emoji":    "😊👍",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var compressedData string
			var err error

			// Call the function with or without compression level
			if len(tt.compressionLevel) > 0 {
				compressedData, err = loader.ObjectsToCompressedJsonLines(tt.objects, tt.compressionLevel[0])
			} else {
				compressedData, err = loader.ObjectsToCompressedJsonLines(tt.objects)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("ObjectsToCompressedJsonLines() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Verify the compressed data is base64-encoded
				_, err := base64.StdEncoding.DecodeString(compressedData)
				if err != nil {
					t.Errorf("Failed to decode base64 data: %v", err)
					return
				}

				// Verify the data is valid gzip
				gzipReader, err := gzip.NewReader(base64Decoder(compressedData))
				if err != nil {
					t.Errorf("Not valid gzip data: %v", err)
					return
				}
				defer gzipReader.Close()

				// For non-empty data, verify we can decompress it and get valid JSON objects
				if len(tt.objects) > 0 {
					var jsonLines string
					buf := make([]byte, 1024)
					for {
						n, err := gzipReader.Read(buf)
						if n > 0 {
							jsonLines += string(buf[:n])
						}
						if err != nil {
							break
						}
					}

					// Verify the number of lines matches the number of objects
					lines := splitLines(jsonLines)
					if len(lines) != len(tt.objects) {
						t.Errorf("Expected %d JSON lines, got %d", len(tt.objects), len(lines))
					}

					// Verify each line is valid JSON
					for i, line := range lines {
						if line == "" {
							continue
						}
						var obj interface{}
						if err := json.Unmarshal([]byte(line), &obj); err != nil {
							t.Errorf("Line %d is not valid JSON: %v", i+1, err)
						}
					}
				}
			}
		})
	}
}

func TestWriteCompressedJsonLinesToArrayFile(t *testing.T) {
	loader := StreamLoader{}
	tempDir := t.TempDir()

	// Create test data: compressed JSON lines
	testObjects := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice"},
		map[string]interface{}{"id": 2, "name": "Bob"},
		map[string]interface{}{"id": 3, "name": "Charlie", "details": map[string]interface{}{"age": 30}},
	}

	compressedJsonLines, err := loader.ObjectsToCompressedJsonLines(testObjects)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	tests := []struct {
		name           string
		compressedData string
		bufferSize     []int
		expectCount    int
		expectError    bool
	}{
		{
			name:           "Valid compressed data",
			compressedData: compressedJsonLines,
			expectCount:    3,
		},
		{
			name:           "With custom buffer size",
			compressedData: compressedJsonLines,
			bufferSize:     []int{128}, // 128 bytes
			expectCount:    3,
		},
		{
			name:           "Invalid base64 data",
			compressedData: "!@#$%^&*()_+",
			expectError:    true,
		},
		{
			name:           "Valid base64 but invalid gzip data",
			compressedData: base64.StdEncoding.EncodeToString([]byte("not gzip data")),
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, "test_"+tt.name+".json")

			var count int
			var err error
			if len(tt.bufferSize) > 0 {
				count, err = loader.WriteCompressedJsonLinesToArrayFile(tt.compressedData, outputPath, tt.bufferSize[0])
			} else {
				count, err = loader.WriteCompressedJsonLinesToArrayFile(tt.compressedData, outputPath)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("WriteCompressedJsonLinesToArrayFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check count is correct
				if count != tt.expectCount {
					t.Errorf("WriteCompressedJsonLinesToArrayFile() count = %v, want %v", count, tt.expectCount)
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

func TestWriteCompressedObjectsToJsonArrayFile(t *testing.T) {
	loader := StreamLoader{}
	tempDir := t.TempDir()

	tests := []struct {
		name             string
		objects          []interface{}
		compressionLevel []int
		expectCount      int
		expectError      bool
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
			name: "With compression level 0 (no compression)",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
			},
			compressionLevel: []int{gzip.NoCompression},
			expectCount:      1,
		},
		{
			name: "With compression level 9 (best compression)",
			objects: []interface{}{
				map[string]interface{}{"id": 1, "name": "Alice"},
			},
			compressionLevel: []int{gzip.BestCompression},
			expectCount:      1,
		},
		{
			name: "Complex objects",
			objects: []interface{}{
				map[string]interface{}{
					"id": 1,
					"profile": map[string]interface{}{
						"name": "Alice",
						"address": map[string]interface{}{
							"street": "123 Main St",
							"city":   "New York",
						},
					},
					"orders": []interface{}{
						map[string]interface{}{"id": "A001", "total": 99.95},
						map[string]interface{}{"id": "A002", "total": 45.50},
					},
				},
			},
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, "compressed_"+tt.name+".json")

			var count int
			var err error
			if len(tt.compressionLevel) > 0 {
				count, err = loader.WriteCompressedObjectsToJsonArrayFile(tt.objects, outputPath, tt.compressionLevel[0])
			} else {
				count, err = loader.WriteCompressedObjectsToJsonArrayFile(tt.objects, outputPath)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("WriteCompressedObjectsToJsonArrayFile() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check count is correct
				if count != tt.expectCount {
					t.Errorf("WriteCompressedObjectsToJsonArrayFile() count = %v, want %v", count, tt.expectCount)
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

					// Compare each object
					for i := 0; i < tt.expectCount; i++ {
						originalBytes, _ := json.Marshal(normalizedOriginal[i])
						resultBytes, _ := json.Marshal(normalizedResult[i])

						if string(originalBytes) != string(resultBytes) {
							t.Errorf("Object at index %d doesn't match after roundtrip", i)
						}
					}
				}
			}
		})
	}
}

func TestCompressedRoundtripFunctionality(t *testing.T) {
	loader := StreamLoader{}
	tempDir := t.TempDir()

	// Test cases for compression roundtrip testing
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
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Convert objects to compressed JSON lines
			compressedData, err := loader.ObjectsToCompressedJsonLines(tt.objects)
			if err != nil {
				t.Fatalf("ObjectsToCompressedJsonLines failed: %v", err)
			}

			// Step 2: Write compressed JSON lines to JSON array file
			jsonArrayFilePath := filepath.Join(tempDir, "compressed_roundtrip_"+tt.name+".json")
			count, err := loader.WriteCompressedJsonLinesToArrayFile(compressedData, jsonArrayFilePath)
			if err != nil {
				t.Fatalf("WriteCompressedJsonLinesToArrayFile failed: %v", err)
			}

			// Verify count matches
			if count != len(tt.objects) {
				t.Errorf("Expected %d objects, got %d", len(tt.objects), count)
			}

			// Step 3: Read back the JSON array and compare with original objects
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

			// Try writing directly with the convenience function
			directWritePath := filepath.Join(tempDir, "direct_compressed_"+tt.name+".json")
			directCount, err := loader.WriteCompressedObjectsToJsonArrayFile(tt.objects, directWritePath)
			if err != nil {
				t.Fatalf("WriteCompressedObjectsToJsonArrayFile failed: %v", err)
			}

			if directCount != len(tt.objects) {
				t.Errorf("Direct write: expected %d objects, got %d", len(tt.objects), directCount)
			}

			// Read the direct write file and verify
			directContent, err := os.ReadFile(directWritePath)
			if err != nil {
				t.Fatalf("Failed to read direct write file: %v", err)
			}

			var directResult []interface{}
			if err := json.Unmarshal(directContent, &directResult); err != nil {
				t.Fatalf("Failed to unmarshal direct write JSON: %v", err)
			}

			if len(directResult) != len(tt.objects) {
				t.Errorf("Direct write object count mismatch: got %d, want %d", len(directResult), len(tt.objects))
			}

			// Compare the two results to verify they're equivalent
			if len(tt.objects) > 0 {
				// Need to normalize representations
				originalJson, _ := json.Marshal(tt.objects)
				resultJson, _ := json.Marshal(result)
				directJson, _ := json.Marshal(directResult)

				var normalizedOriginal, normalizedResult, normalizedDirect []interface{}
				json.Unmarshal(originalJson, &normalizedOriginal)
				json.Unmarshal(resultJson, &normalizedResult)
				json.Unmarshal(directJson, &normalizedDirect)

				// Compare multi-step and direct write results
				for i := 0; i < len(normalizedOriginal); i++ {
					resultBytes, _ := json.Marshal(normalizedResult[i])
					directBytes, _ := json.Marshal(normalizedDirect[i])

					if string(resultBytes) != string(directBytes) {
						t.Errorf("Multi-step and direct write results don't match for object %d", i)
					}
				}
			}
		})
	}
}

// Helper function to create a base64 decoder
func base64Decoder(encoded string) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))
}
