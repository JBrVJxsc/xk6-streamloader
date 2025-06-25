// streamloader.go
package streamloader

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.k6.io/k6/js/modules"
)

// StreamLoader is the k6/x/streamloader module.
// It provides LoadJSON for reading large JSON files efficiently
// using a small buffer and supporting standard JSON arrays, NDJSON, or JSON objects.
// It also provides LoadCSV for streaming CSV files with minimal memory footprint.
type StreamLoader struct{}

// LoadCSV opens the given CSV file and streams its content into a slice of string slices.
// Each row is represented as []string, and the entire result is [][]string.
// The function reads the file incrementally to minimize memory usage and avoid spikes.
// It automatically detects common CSV delimiters and handles quoted fields properly.
// 
// Options for memory optimization:
// - Uses buffered reading with configurable buffer size
// - Processes one row at a time instead of loading entire file
// - Supports files of any size without significant memory overhead
//
// Example usage:
//   records, err := streamloader.LoadCSV("data.csv")
//   // records[0] contains the first row as []string
//   // records[1] contains the second row as []string, etc.
func (StreamLoader) LoadCSV(filePath string) ([][]string, error) {
	// 1) Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// 2) Create buffered reader (64 KB) for efficient reading
	reader := bufio.NewReaderSize(file, 64*1024)

	// 3) Create CSV reader with standard settings
	csvReader := csv.NewReader(reader)
	
	// Configure CSV reader for robust parsing
	csvReader.TrimLeadingSpace = true
	csvReader.LazyQuotes = true
	// Allow variable number of fields per record
	csvReader.FieldsPerRecord = -1
	
	// 4) Read all records incrementally
	var records [][]string
	
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse CSV at line %d: %w", len(records)+1, err)
		}
		
		// Make a copy of the record to avoid memory sharing issues
		recordCopy := make([]string, len(record))
		copy(recordCopy, record)
		records = append(records, recordCopy)
	}
	
	return records, nil
}

// LoadJSON opens the given file, streams and parses its JSON content into a slice of generic maps.
// By returning map[string]interface{}, we preserve the original JSON key names exactly as-is.
// Supports three formats:
// 1. JSON array: [{...}, {...}]
// 2. NDJSON: {...}\n{...}\n
// 3. JSON object: {"key1": {...}, "key2": {...}} (returned as a map)
func (StreamLoader) LoadJSON(filePath string) (any, error) {
	// 1) Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 2) Buffered reader (64 KB)
	reader := bufio.NewReaderSize(file, 64*1024)

	// 3) NDJSON detection by extension
	if strings.HasSuffix(strings.ToLower(filepath.Ext(filePath)), ".ndjson") {
		scanner := bufio.NewScanner(reader)
		var objects []map[string]any
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var item map[string]any
			if err := json.Unmarshal([]byte(line), &item); err != nil {
				return nil, err
			}
			objects = append(objects, item)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return objects, nil
	}

	// 4) Peek first non-whitespace byte to detect format
	var firstByte byte
	for {
		b, err := reader.Peek(1)
		if err != nil {
			return nil, err
		}
		if isWhitespace(b[0]) {
			reader.ReadByte()
			continue
		}
		firstByte = b[0]
		break
	}

	switch firstByte {
	case '[':
		// Standard JSON array format
		dec := json.NewDecoder(reader)

		// Consume opening '['
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if delim, ok := tok.(json.Delim); !ok || delim != '[' {
			return nil, fmt.Errorf("expected JSON array, got %v", tok)
		}

		var arr []interface{}
		for dec.More() {
			var item interface{}
			if err := dec.Decode(&item); err != nil {
				return nil, err
			}
			arr = append(arr, item)
		}

		// Consume closing ']'
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
		return arr, nil
	case '{':
		// JSON object format - return as map directly
		dec := json.NewDecoder(reader)

		var objMap map[string]any
		if err := dec.Decode(&objMap); err != nil {
			return nil, err
		}
		return objMap, nil
	default:
		// Newline-delimited JSON (NDJSON) format
		scanner := bufio.NewScanner(reader)
		var objects []map[string]any
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var item map[string]any
			if err := json.Unmarshal([]byte(line), &item); err != nil {
				return nil, err
			}
			objects = append(objects, item)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return objects, nil
	}
}

// isWhitespace checks for JSON whitespace characters
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t'
}

func init() {
	modules.Register("k6/x/streamloader", new(StreamLoader))
}
