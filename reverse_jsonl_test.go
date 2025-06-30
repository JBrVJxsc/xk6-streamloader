package streamloader_test

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"

	. "github.com/JBrVJxsc/xk6-streamloader"
)

func TestJsonLinesToObjects(t *testing.T) {
	loader := StreamLoader{}

	tests := []struct {
		name        string
		jsonLines   string
		expectCount int
		expectError bool
	}{
		{
			name:        "Empty string",
			jsonLines:   "",
			expectCount: 0,
		},
		{
			name:        "Single object",
			jsonLines:   `{"id":1,"name":"Alice"}`,
			expectCount: 1,
		},
		{
			name:        "Multiple objects",
			jsonLines:   `{"id":1,"name":"Alice"}
{"id":2,"name":"Bob"}
{"id":3,"name":"Charlie"}`,
			expectCount: 3,
		},
		{
			name:        "With blank lines",
			jsonLines:   `{"id":1,"name":"Alice"}

{"id":2,"name":"Bob"}

`,
			expectCount: 2,
		},
		{
			name:        "With whitespace",
			jsonLines:   `  {"id":1,"name":"Alice"}  
  {"id":2,"name":"Bob"}  `,
			expectCount: 2,
		},
		{
			name:        "Complex objects",
			jsonLines:   `{"id":1,"name":"Alice","details":{"age":30,"city":"New York"}}
{"id":2,"name":"Bob","skills":["Java","Python"]}`,
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
			objects, err := loader.JsonLinesToObjects(tt.jsonLines)
			if (err != nil) != tt.expectError {
				t.Errorf("JsonLinesToObjects() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check count is correct
				if len(objects) != tt.expectCount {
					t.Errorf("JsonLinesToObjects() count = %v, want %v", len(objects), tt.expectCount)
				}
			}
		})
	}
}

func TestCompressedJsonLinesToObjects(t *testing.T) {
	loader := StreamLoader{}

	// Create test objects for compression
	testObjects := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice"},
		map[string]interface{}{"id": 2, "name": "Bob"},
		map[string]interface{}{"id": 3, "name": "Charlie", "details": map[string]interface{}{"age": 30}},
	}

	// Compress the objects to get compressed JSON lines
	compressedJsonLines, err := loader.ObjectsToCompressedJsonLines(testObjects)
	if err != nil {
		t.Fatalf("Failed to create test data: %v", err)
	}

	// Create test cases
	tests := []struct {
		name           string
		compressedData string
		expectCount    int
		expectError    bool
	}{
		{
			name:           "Valid compressed data",
			compressedData: compressedJsonLines,
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
			objects, err := loader.CompressedJsonLinesToObjects(tt.compressedData)
			if (err != nil) != tt.expectError {
				t.Errorf("CompressedJsonLinesToObjects() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				// Check count is correct
				if len(objects) != tt.expectCount {
					t.Errorf("CompressedJsonLinesToObjects() count = %v, want %v", len(objects), tt.expectCount)
				}
			}
		})
	}
}

func TestReverseRoundtripFunctionality(t *testing.T) {
	loader := StreamLoader{}

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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Step 1: Convert objects to JSON lines
			jsonLines, err := loader.ObjectsToJsonLines(tc.objects)
			if err != nil {
				t.Fatalf("ObjectsToJsonLines failed: %v", err)
			}

			// Step 2: Convert JSON lines back to objects using the new function
			parsedObjects, err := loader.JsonLinesToObjects(jsonLines)
			if err != nil {
				t.Fatalf("JsonLinesToObjects failed: %v", err)
			}

			// Step 3: Compare original and parsed objects
			if len(parsedObjects) != len(tc.objects) {
				t.Errorf("Object count mismatch: got %d, want %d", len(parsedObjects), len(tc.objects))
			}

			// Need to normalize both object representations by marshaling to JSON and back
			originalJson, _ := json.Marshal(tc.objects)
			parsedJson, _ := json.Marshal(parsedObjects)

			var normalizedOriginal, normalizedParsed []interface{}
			json.Unmarshal(originalJson, &normalizedOriginal)
			json.Unmarshal(parsedJson, &normalizedParsed)

			if !reflect.DeepEqual(normalizedOriginal, normalizedParsed) {
				t.Errorf("Objects don't match after roundtrip.\nOriginal: %v\nParsed: %v", normalizedOriginal, normalizedParsed)
			}

			// Step 4: Test compressed roundtrip
			compressedJsonLines, err := loader.ObjectsToCompressedJsonLines(tc.objects)
			if err != nil {
				t.Fatalf("ObjectsToCompressedJsonLines failed: %v", err)
			}

			// Step 5: Convert compressed JSON lines back to objects
			compressedObjects, err := loader.CompressedJsonLinesToObjects(compressedJsonLines)
			if err != nil {
				t.Fatalf("CompressedJsonLinesToObjects failed: %v", err)
			}

			// Step 6: Compare original and decompressed objects
			if len(compressedObjects) != len(tc.objects) {
				t.Errorf("Compressed object count mismatch: got %d, want %d", len(compressedObjects), len(tc.objects))
			}

			// Need to normalize both object representations by marshaling to JSON and back
			compressedJson, _ := json.Marshal(compressedObjects)

			var normalizedCompressed []interface{}
			json.Unmarshal(compressedJson, &normalizedCompressed)

			if !reflect.DeepEqual(normalizedOriginal, normalizedCompressed) {
				t.Errorf("Objects don't match after compression roundtrip.\nOriginal: %v\nCompressed: %v", normalizedOriginal, normalizedCompressed)
			}
		})
	}
}

func TestMultiStepRoundtrip(t *testing.T) {
	loader := StreamLoader{}

	// Create test objects
	objects := []interface{}{
		map[string]interface{}{"id": 1, "name": "Alice"},
		map[string]interface{}{"id": 2, "name": "Bob"},
		map[string]interface{}{
			"id": 3, 
			"name": "Charlie",
			"details": map[string]interface{}{
				"age": 30,
				"city": "New York",
			},
		},
	}

	// Step 1: Convert objects to JSON lines
	jsonLines, err := loader.ObjectsToJsonLines(objects)
	if err != nil {
		t.Fatalf("ObjectsToJsonLines failed: %v", err)
	}

	// Step 2: Convert JSON lines back to objects
	parsedObjects, err := loader.JsonLinesToObjects(jsonLines)
	if err != nil {
		t.Fatalf("JsonLinesToObjects failed: %v", err)
	}

	// Step 3: Convert objects to compressed JSON lines
	compressedJsonLines, err := loader.ObjectsToCompressedJsonLines(parsedObjects)
	if err != nil {
		t.Fatalf("ObjectsToCompressedJsonLines failed: %v", err)
	}

	// Step 4: Convert compressed JSON lines back to objects
	finalObjects, err := loader.CompressedJsonLinesToObjects(compressedJsonLines)
	if err != nil {
		t.Fatalf("CompressedJsonLinesToObjects failed: %v", err)
	}

	// Compare original and final objects

	if len(finalObjects) != len(objects) {
		t.Errorf("Object count mismatch: got %d, want %d", len(finalObjects), len(objects))
	}

	// Need to normalize both object representations
	originalJson, _ := json.Marshal(objects)
	finalObjectsJson, _ := json.Marshal(finalObjects)

	var normalizedOriginal, normalizedFinal []interface{}
	json.Unmarshal(originalJson, &normalizedOriginal)
	json.Unmarshal(finalObjectsJson, &normalizedFinal)

	if !reflect.DeepEqual(normalizedOriginal, normalizedFinal) {
		t.Errorf("Objects don't match after multi-step roundtrip.")
		t.Errorf("Original: %v", normalizedOriginal)
		t.Errorf("Final: %v", normalizedFinal)
	}
}

func TestEdgeCases(t *testing.T) {
	loader := StreamLoader{}
	
	// Test handling of empty strings
	objects, err := loader.JsonLinesToObjects("")
	if err != nil {
		t.Errorf("JsonLinesToObjects should handle empty string: %v", err)
	}
	if len(objects) != 0 {
		t.Errorf("JsonLinesToObjects with empty string should return empty array, got %d items", len(objects))
	}
	
	// Test handling of whitespace-only string
	objects, err = loader.JsonLinesToObjects("   \n   \t   ")
	if err != nil {
		t.Errorf("JsonLinesToObjects should handle whitespace-only string: %v", err)
	}
	if len(objects) != 0 {
		t.Errorf("JsonLinesToObjects with whitespace-only string should return empty array, got %d items", len(objects))
	}
	
	// Test handling of mixed valid/invalid JSON
	_, err = loader.JsonLinesToObjects(`{"id":1}\ninvalid\n{"id":2}`)
	if err == nil {
		t.Errorf("JsonLinesToObjects should return error with invalid JSON")
	}
	
	// Test handling of invalid base64 for compressed function
	_, err = loader.CompressedJsonLinesToObjects("not-valid-base64")
	if err == nil {
		t.Errorf("CompressedJsonLinesToObjects should return error with invalid base64")
	}
	
	// Test handling of valid base64 but invalid gzip
	invalidGzip := base64.StdEncoding.EncodeToString([]byte("not-gzip-data"))
	_, err = loader.CompressedJsonLinesToObjects(invalidGzip)
	if err == nil {
		t.Errorf("CompressedJsonLinesToObjects should return error with invalid gzip data")
	}
}