// streamloader.go
package streamloader

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"container/ring"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"go.k6.io/k6/js/modules"
)

// StreamLoader is the k6/x/streamloader module.
// It provides LoadJSON for reading large JSON files efficiently
// using a small buffer and supporting standard JSON arrays, NDJSON, or JSON objects.
// It also provides LoadCSV for streaming CSV files with minimal memory footprint.
// Additionally, it includes utilities for converting between JSON formats and working with compressed JSON data.
type StreamLoader struct{}

// FilterConfig represents a row filter configuration
type FilterConfig struct {
	Type    string   `json:"type" js:"type"`
	Column  int      `json:"column" js:"column"`
	Pattern string   `json:"pattern,omitempty" js:"pattern"`
	Min     *float64 `json:"min,omitempty" js:"min"`
	Max     *float64 `json:"max,omitempty" js:"max"`
}

// TransformConfig represents a value transform configuration
type TransformConfig struct {
	Type   string      `json:"type" js:"type"`
	Column int         `json:"column" js:"column"`
	Value  interface{} `json:"value,omitempty" js:"value"`
	Start  int         `json:"start,omitempty" js:"start"`
	Length *int        `json:"length,omitempty" js:"length"`
}

// GroupByConfig represents grouping configuration
type GroupByConfig struct {
	Column int `json:"column" js:"column"`
}

// FieldConfig represents a projection field configuration
type FieldConfig struct {
	Type   string      `json:"type" js:"type"`
	Column int         `json:"column,omitempty" js:"column"`
	Value  interface{} `json:"value,omitempty" js:"value"`
}

// CsvOptions represents options for CSV parsing in LoadCSV
type CsvOptions struct {
	LazyQuotes       bool `json:"lazyQuotes" js:"lazyQuotes"`
	TrimLeadingSpace bool `json:"trimLeadingSpace" js:"trimLeadingSpace"`
	TrimSpace        bool `json:"trimSpace" js:"trimSpace"`
	ReuseRecord      bool `json:"reuseRecord" js:"reuseRecord"`
}

// ProcessCsvOptions represents options for ProcessCsvFile
type ProcessCsvOptions struct {
	SkipHeader       bool              `json:"skipHeader" js:"skipHeader"`
	LazyQuotes       bool              `json:"lazyQuotes" js:"lazyQuotes"`
	TrimLeadingSpace bool              `json:"trimLeadingSpace" js:"trimLeadingSpace"`
	TrimSpace        bool              `json:"trimSpace" js:"trimSpace"`
	ReuseRecord      bool              `json:"reuseRecord" js:"reuseRecord"`
	Filters          []FilterConfig    `json:"filters" js:"filters"`
	Transforms       []TransformConfig `json:"transforms" js:"transforms"`
	GroupBy          *GroupByConfig    `json:"groupBy,omitempty" js:"groupBy"`
	Fields           []FieldConfig     `json:"fields" js:"fields"`
}

// ProcessCsvFile opens a CSV file and processes it row by row using streaming to minimize memory usage.
// It applies filters, transforms, grouping, and projection in a single pass through the file.
// This approach is memory-efficient for large CSV files since it processes one row at a time
// instead of loading the entire file into memory first.
//
// Options:
// - skipHeader: Whether to skip the first row as header (default: true)
// - lazyQuotes: Allow unescaped quotes in quoted fields (default: true)
// - trimLeadingSpace: Trim leading whitespace from fields (default: true)
// - trimSpace: Trim all whitespace from fields (leading and trailing) (default: false)
// - reuseRecord: Reuse record memory for better performance (default: true)
// - filters: Array of filter configs to drop unwanted rows:
//   - { type: "emptyString", column: N }
//   - { type: "regexMatch", column: N, pattern: "regex" }
//   - { type: "valueRange", column: N, min: X, max: Y }
//
// - transforms: Array of transform configs to apply in-place:
//   - { type: "parseInt", column: N }
//   - { type: "fixedValue", column: N, value: V }
//   - { type: "substring", column: N, start: S, length: L }
//
// - groupBy: Optional grouping by column: { column: N }
// - fields: Projection fields:
//   - { type: "column", column: N } | { type: "fixed", value: V }
//
// Returns: Array of arrays containing processed data, grouped if groupBy is specified
//
// Example usage:
//
//	options := ProcessCsvOptions{
//		SkipHeader: true,
//		Filters: []FilterConfig{
//			{Type: "emptyString", Column: 0},
//		},
//		Transforms: []TransformConfig{
//			{Type: "parseInt", Column: 1},
//		},
//		Fields: []FieldConfig{
//			{Type: "column", Column: 0},
//			{Type: "column", Column: 1},
//		},
//	}
//	result, err := streamloader.ProcessCsvFile("data.csv", options)
func (StreamLoader) ProcessCsvFile(filePath string, options ProcessCsvOptions) ([][]interface{}, error) {
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
	csvReader.TrimLeadingSpace = true // Default to true
	if !options.TrimLeadingSpace {    // Only override if explicitly set to false
		csvReader.TrimLeadingSpace = false
	}
	csvReader.LazyQuotes = options.LazyQuotes // Use configurable setting
	// Allow variable number of fields per record
	csvReader.FieldsPerRecord = -1
	// Apply ReuseRecord option with default true for performance
	csvReader.ReuseRecord = true // Default to true
	if !options.ReuseRecord {    // Only override if explicitly set to false
		csvReader.ReuseRecord = false
	}

	// 4) Initialize processing state
	var rowIndex int
	skipHeader := options.SkipHeader
	hasGrouping := options.GroupBy != nil
	var groupMap map[string][][]interface{}
	var result [][]interface{}

	if hasGrouping {
		groupMap = make(map[string][][]interface{})
	}

	// Pre-compile regex patterns for performance
	regexCache := make(map[string]*regexp.Regexp)
	for _, filter := range options.Filters {
		if filter.Type == "regexMatch" {
			compiled, err := regexp.Compile(filter.Pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex pattern in filter: %w", err)
			}
			regexCache[filter.Pattern] = compiled
		}
	}

	// 5) Process rows one by one
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse CSV at line %d: %w", rowIndex+1, err)
		}

		// Skip header if requested
		if rowIndex == 0 && skipHeader {
			rowIndex++
			continue
		}

		// Make a copy and normalize fields
		row := make([]string, len(record))

		// Apply trimming according to options
		if options.TrimSpace {
			for i, field := range record {
				// Trim all whitespace
				row[i] = strings.TrimSpace(field)
			}
		} else {
			// Just copy if no trimming required
			copy(row, record)
		}

		// Apply filters
		shouldDrop := false
		for _, filter := range options.Filters {
			if filter.Column >= len(row) {
				shouldDrop = true
				break // Drop the row if column doesn't exist
			}

			cell := row[filter.Column]
			switch filter.Type {
			case "emptyString":
				if cell == "" {
					shouldDrop = true
				}
			case "regexMatch":
				if regex, exists := regexCache[filter.Pattern]; exists {
					if !regex.MatchString(cell) {
						shouldDrop = true
					}
				}
			case "valueRange":
				if num, err := strconv.ParseFloat(cell, 64); err == nil {
					if (filter.Min != nil && num < *filter.Min) ||
						(filter.Max != nil && num > *filter.Max) {
						shouldDrop = true
					}
				} else {
					// Treat non-numeric values as not satisfying the range
					shouldDrop = true
				}
			}
			if shouldDrop {
				break
			}
		}

		if shouldDrop {
			rowIndex++
			continue
		}

		// Apply transforms
		for _, transform := range options.Transforms {
			if transform.Column >= len(row) {
				continue // Skip transform if column doesn't exist
			}

			switch transform.Type {
			case "parseInt":
				if num, err := strconv.Atoi(row[transform.Column]); err == nil {
					row[transform.Column] = fmt.Sprintf("%d", num)
				}
			case "fixedValue":
				row[transform.Column] = fmt.Sprintf("%v", transform.Value)
			case "substring":
				str := row[transform.Column]
				start := transform.Start
				if start < 0 || start >= len(str) {
					row[transform.Column] = ""
				} else {
					end := len(str)
					if transform.Length != nil && *transform.Length > 0 {
						if start+*transform.Length < len(str) {
							end = start + *transform.Length
						}
					}
					row[transform.Column] = str[start:end]
				}
			}
		}

		// Build projected row
		var projected []interface{}
		if len(options.Fields) > 0 {
			for _, field := range options.Fields {
				switch field.Type {
				case "column":
					if field.Column < len(row) {
						projected = append(projected, row[field.Column])
					} else {
						projected = append(projected, "")
					}
				case "fixed":
					projected = append(projected, field.Value)
				}
			}
		} else {
			// If no fields are specified, project all columns as strings
			for _, col := range row {
				projected = append(projected, col)
			}
		}

		// Handle grouping or direct collection
		if hasGrouping {
			if options.GroupBy.Column < len(row) {
				key := row[options.GroupBy.Column]
				if groupMap[key] == nil {
					groupMap[key] = make([][]interface{}, 0)
				}
				groupMap[key] = append(groupMap[key], projected)
			}
		} else {
			result = append(result, projected)
		}

		rowIndex++
	}

	// 7) Finalize output
	if hasGrouping {
		// Convert grouped data to flat arrays
		groupedResult := make([][]interface{}, 0, len(groupMap))
		for _, group := range groupMap {
			// Flatten each group into a single array
			var flatGroup []interface{}
			for _, row := range group {
				flatGroup = append(flatGroup, row...)
			}
			groupedResult = append(groupedResult, flatGroup)
		}
		return groupedResult, nil
	}

	return result, nil
}

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
// Available options:
// - lazyQuotes: Controls how strictly the CSV parser handles quotes (default: true)
//   - When true, quotes may appear in an unquoted field and a non-doubled quote may appear in a quoted field
//   - When false, strict RFC 4180 compliance is enforced
//
// - trimLeadingSpace: Removes leading whitespace from fields (default: true)
//   - Only removes whitespace at the beginning of fields when a field is read
//   - This is a built-in feature of the CSV reader
//
// - trimSpace: Removes all whitespace from fields (leading and trailing) (default: false)
//   - This is a more aggressive trimming that removes both leading and trailing whitespace
//   - When true, this overrides the CSV reader's built-in TrimLeadingSpace behavior
//
// - reuseRecord: Reuses record memory for better performance (default: true)
//   - Reduces memory allocations by reusing the same slice for each record
//   - Only set to false if you need to retain references to individual records
//
// Example usage:
//
// With detailed options:
//
//	options := CsvOptions{
//	    LazyQuotes: true,
//	    TrimLeadingSpace: true,
//	    TrimSpace: false,
//	    ReuseRecord: true,
//	}
//	records, err := streamloader.LoadCSV("data.csv", options)
//
// With simple boolean for backward compatibility (lazy quotes):
//
//	records, err := streamloader.LoadCSV("data.csv", true) // With lazy quotes
//
// With default settings (all options true except TrimSpace):
//
//	records, err := streamloader.LoadCSV("data.csv")
//
//	// records[0] contains the first row as []string
//	// records[1] contains the second row as []string, etc.
func (s StreamLoader) LoadCSV(filePath string, options ...interface{}) ([][]string, error) {
	// Set defaults
	isLazyQuotes := true
	isTrimLeadingSpace := true
	isTrimSpace := false
	isReuseRecord := true

	// Process options if provided
	if len(options) > 0 {
		// First try to process as CsvOptions struct
		if csvOptions, ok := options[0].(CsvOptions); ok {
			isLazyQuotes = csvOptions.LazyQuotes
			isTrimLeadingSpace = csvOptions.TrimLeadingSpace
			isTrimSpace = csvOptions.TrimSpace
			isReuseRecord = csvOptions.ReuseRecord
		} else if lazyQuotes, ok := options[0].(bool); ok {
			// Backward compatibility: interpret bool as LazyQuotes
			isLazyQuotes = lazyQuotes
		}
	}
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
	csvReader.TrimLeadingSpace = isTrimLeadingSpace
	csvReader.LazyQuotes = isLazyQuotes
	// Allow variable number of fields per record
	csvReader.FieldsPerRecord = -1
	// Apply ReuseRecord option for memory efficiency
	csvReader.ReuseRecord = isReuseRecord

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

		// Apply TrimSpace if enabled
		if isTrimSpace {
			for i, field := range record {
				recordCopy[i] = strings.TrimSpace(field)
			}
		} else {
			copy(recordCopy, record)
		}

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

// LoadText opens the given file and reads its entire content into a string.
// This function is optimized for performance and is suitable for loading moderate-sized text files.
// It uses os.ReadFile for an efficient single-read operation.
//
// Example usage:
//
//	content, err := streamloader.LoadText("data.txt")
func (StreamLoader) LoadText(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(bytes), nil
}

// Head reads the first N lines of a file without loading the entire file into memory.
// It returns the lines as a single string, with each line separated by a newline character.
// This is useful for previewing large files without consuming excessive memory.
//
// Example usage:
//
//	first10Lines, err := streamloader.Head("large_file.txt", 10)
func (StreamLoader) Head(filePath string, n int) (string, error) {
	if n <= 0 {
		return "", nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for i := 0; i < n && scanner.Scan(); i++ {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	if len(lines) == 0 {
		return "", nil
	}

	return strings.Join(lines, "\n"), nil
}

// Tail reads the last N lines of a file without loading the entire file into memory.
// It returns the lines as a single string, with each line separated by a newline character.
// This is useful for previewing the end of large files.
//
// Example usage:
//
//	last10Lines, err := streamloader.Tail("large_file.txt", 10)
func (StreamLoader) Tail(filePath string, n int) (string, error) {
	if n <= 0 {
		return "", nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	ringBuffer := ring.New(n)
	for scanner.Scan() {
		ringBuffer.Value = scanner.Text()
		ringBuffer = ringBuffer.Next()
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	var resultLines []string
	ringBuffer.Do(func(p interface{}) {
		if p != nil {
			resultLines = append(resultLines, p.(string))
		}
	})

	return strings.Join(resultLines, "\n"), nil
}

// isWhitespace checks for JSON whitespace characters
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t'
}

// DebugOptions returns the options structure exactly as received for debugging parameter passing
func (StreamLoader) DebugOptions(options interface{}) interface{} {
	return options
}

// DebugCsvOptions returns the ProcessCsvOptions exactly as received for debugging parameter passing
func (StreamLoader) DebugCsvOptions(options ProcessCsvOptions) ProcessCsvOptions {
	return options
}

// ObjectsToJsonLines converts a slice of JavaScript objects (represented as maps) into JSONL format.
// Each object is JSON-encoded and placed on a separate line with a newline character separator.
// This is useful for efficiently serializing large datasets for storage or streaming.
//
// Parameters:
//   - objects: An array of JavaScript objects to convert to JSONL format.
//
// Returns:
//   - A string containing the JSONL representation of the objects.
//
// Example:
//
//	objects = [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]
//	jsonLines = streamloader.ObjectsToJsonLines(objects)
//	// jsonLines will be '{"id":1,"name":"Alice"}\n{"id":2,"name":"Bob"}'
func (StreamLoader) ObjectsToJsonLines(objects []interface{}) (string, error) {
	var builder strings.Builder
	encoder := json.NewEncoder(&builder)
	encoder.SetEscapeHTML(false) // Avoid escaping HTML entities like &, <, >

	for i, obj := range objects {
		if err := encoder.Encode(obj); err != nil {
			return "", fmt.Errorf("failed to encode object at index %d: %w", i, err)
		}
	}

	// The encoder adds a newline after each object, which is what we want for JSONL format
	// We just need to trim the trailing newline if present
	jsonLines := builder.String()
	if len(jsonLines) > 0 && jsonLines[len(jsonLines)-1] == '\n' {
		jsonLines = jsonLines[:len(jsonLines)-1]
	}

	return jsonLines, nil
}

// ObjectsToCompressedJsonLines converts a slice of JavaScript objects into JSONL format and
// compresses the result using gzip. The compressed data is then base64-encoded to make it
// easy to transport as a string. This is useful for efficiently serializing and compressing
// large datasets.
//
// Parameters:
//   - objects: An array of JavaScript objects to convert to compressed JSONL format.
//   - compressionLevel: Optional compression level (0-9, where 0=no compression, 1=best speed,
//     9=best compression). Default is gzip.DefaultCompression (-1).
//
// Returns:
//   - A base64-encoded string containing the gzip-compressed JSONL data.
//
// Example:
//
//	objects = [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]
//	compressedJsonLines = streamloader.ObjectsToCompressedJsonLines(objects)
//	// Returns base64-encoded gzipped JSON lines
func (s StreamLoader) ObjectsToCompressedJsonLines(objects []interface{}, compressionLevel ...int) (string, error) {
	// First convert objects to JSON lines
	jsonLines, err := s.ObjectsToJsonLines(objects)
	if err != nil {
		return "", fmt.Errorf("failed to convert objects to JSON lines: %w", err)
	}

	// Set default compression level if not provided
	level := gzip.DefaultCompression
	if len(compressionLevel) > 0 && compressionLevel[0] >= gzip.NoCompression && compressionLevel[0] <= gzip.BestCompression {
		level = compressionLevel[0]
	}

	// Compress the JSON lines with gzip
	var compressedBuffer bytes.Buffer
	gzWriter, err := gzip.NewWriterLevel(&compressedBuffer, level)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip writer: %w", err)
	}

	// Write the JSON lines to the gzip writer
	if _, err := gzWriter.Write([]byte(jsonLines)); err != nil {
		gzWriter.Close()
		return "", fmt.Errorf("failed to compress data: %w", err)
	}

	// Close the gzip writer to flush all data
	if err := gzWriter.Close(); err != nil {
		return "", fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// Base64 encode the compressed data
	compressedBase64 := base64.StdEncoding.EncodeToString(compressedBuffer.Bytes())
	return compressedBase64, nil
}

// WriteJsonLinesToArrayFile reads JSONL-formatted data (one JSON object per line) and writes it
// as a single JSON array to a file. It streams the output to minimize memory usage, making it
// suitable for very large datasets.
//
// Parameters:
//   - jsonLines: A string containing JSONL-formatted data, with one JSON object per line.
//   - outputFilePath: The path where the resulting JSON array file will be written.
//   - bufferSize: Optional buffer size in bytes (default: 64KB). Determines how much data is
//     buffered before writing to disk.
//
// Returns:
//   - The count of objects written to the file.
//   - An error if the operation failed.
//
// Example:
//
//	jsonLines := '{"id":1,"name":"Alice"}\n{"id":2,"name":"Bob"}'
//	count, err := streamloader.WriteJsonLinesToArrayFile(jsonLines, "output.json")
//	// Will write '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]' to output.json
func (StreamLoader) WriteJsonLinesToArrayFile(jsonLines string, outputFilePath string, bufferSize ...int) (int, error) {
	// Set default buffer size if not provided
	bufSize := 64 * 1024 // 64KB default
	if len(bufferSize) > 0 && bufferSize[0] > 0 {
		bufSize = bufferSize[0]
	}

	// Create or truncate the output file
	file, err := os.Create(outputFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create a buffered writer for efficiency
	writer := bufio.NewWriterSize(file, bufSize)
	defer writer.Flush()

	// Write the opening bracket of the JSON array
	if _, err := writer.WriteString("["); err != nil {
		return 0, fmt.Errorf("failed to write opening bracket: %w", err)
	}

	// Process the JSON lines
	scanner := bufio.NewScanner(strings.NewReader(jsonLines))
	// For very large lines, increase the scanner buffer size
	scanner.Buffer(make([]byte, bufSize), 10*bufSize)

	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip empty lines
		}

		// Write comma separator for all but the first object
		if count > 0 {
			if _, err := writer.WriteString(","); err != nil {
				return count, fmt.Errorf("failed to write comma separator: %w", err)
			}
		}

		// Validate that the line is a valid JSON object
		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			return count, fmt.Errorf("invalid JSON at line %d: %w", count+1, err)
		}

		// Write the JSON object to the file
		if _, err := writer.WriteString(line); err != nil {
			return count, fmt.Errorf("failed to write JSON object: %w", err)
		}

		count++
	}

	if err := scanner.Err(); err != nil {
		return count, fmt.Errorf("error reading JSON lines: %w", err)
	}

	// Write the closing bracket of the JSON array
	if _, err := writer.WriteString("]"); err != nil {
		return count, fmt.Errorf("failed to write closing bracket: %w", err)
	}

	// Flush any buffered data to the file
	if err := writer.Flush(); err != nil {
		return count, fmt.Errorf("failed to flush data to file: %w", err)
	}

	return count, nil
}

// WriteCompressedJsonLinesToArrayFile decompresses gzipped, base64-encoded JSONL data and writes
// it as a single JSON array to a file. It streams the output to minimize memory usage, making it
// suitable for very large compressed datasets.
//
// Parameters:
//   - compressedJsonLines: A base64-encoded string containing gzip-compressed JSONL data.
//   - outputFilePath: The path where the resulting JSON array file will be written.
//   - bufferSize: Optional buffer size in bytes (default: 64KB). Determines how much data is
//     buffered before writing to disk.
//
// Returns:
//   - The count of objects written to the file.
//   - An error if the operation failed.
//
// Example:
//
//	compressedData := "H4sIAAAAAAAA/6tWSk5OLCpKVbJSMjA2M9RRKsgsVrIyBHITKzNSixQUQPLJ..."
//	count, err := streamloader.WriteCompressedJsonLinesToArrayFile(compressedData, "output.json")
//	// Will decompress and write the JSON array to output.json
func (StreamLoader) WriteCompressedJsonLinesToArrayFile(compressedJsonLines string, outputFilePath string, bufferSize ...int) (int, error) {
	// Set default buffer size if not provided
	bufSize := 64 * 1024 // 64KB default
	if len(bufferSize) > 0 && bufferSize[0] > 0 {
		bufSize = bufferSize[0]
	}

	// Decode base64 data
	compressedData, err := base64.StdEncoding.DecodeString(compressedJsonLines)
	if err != nil {
		return 0, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Set up the gzip reader to decompress the data
	gzReader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return 0, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create or truncate the output file
	file, err := os.Create(outputFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create a buffered writer for efficiency
	writer := bufio.NewWriterSize(file, bufSize)
	defer writer.Flush()

	// Write the opening bracket of the JSON array
	if _, err := writer.WriteString("["); err != nil {
		return 0, fmt.Errorf("failed to write opening bracket: %w", err)
	}

	// Process the decompressed JSON lines
	scanner := bufio.NewScanner(gzReader)
	// For very large lines, increase the scanner buffer size
	scanner.Buffer(make([]byte, bufSize), 10*bufSize)

	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip empty lines
		}

		// Write comma separator for all but the first object
		if count > 0 {
			if _, err := writer.WriteString(","); err != nil {
				return count, fmt.Errorf("failed to write comma separator: %w", err)
			}
		}

		// Validate that the line is a valid JSON object
		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			return count, fmt.Errorf("invalid JSON at line %d: %w", count+1, err)
		}

		// Write the JSON object to the file
		if _, err := writer.WriteString(line); err != nil {
			return count, fmt.Errorf("failed to write JSON object: %w", err)
		}

		count++
	}

	if err := scanner.Err(); err != nil {
		return count, fmt.Errorf("error reading decompressed JSON lines: %w", err)
	}

	// Write the closing bracket of the JSON array
	if _, err := writer.WriteString("]"); err != nil {
		return count, fmt.Errorf("failed to write closing bracket: %w", err)
	}

	// Flush any buffered data to the file
	if err := writer.Flush(); err != nil {
		return count, fmt.Errorf("failed to flush data to file: %w", err)
	}

	return count, nil
}

// CombineJsonArrayFiles combines multiple JSON array files into a single JSON array file.
// This is useful for merging data from multiple sources or processing large datasets in chunks.
// It streams the data to minimize memory usage, making it suitable for very large files.
//
// Parameters:
//   - inputFilePaths: An array of paths to JSON array files to combine.
//   - outputFilePath: The path where the resulting combined JSON array will be written.
//   - bufferSize: Optional buffer size in bytes (default: 64KB).
//
// Returns:
//   - The count of objects written to the file.
//   - An error if the operation failed.
//
// Example:
//
//	count, err := streamloader.CombineJsonArrayFiles(["file1.json", "file2.json"], "combined.json")
//	// Will merge the arrays from file1.json and file2.json into combined.json
func (StreamLoader) CombineJsonArrayFiles(inputFilePaths []string, outputFilePath string, bufferSize ...int) (int, error) {
	// Set default buffer size if not provided
	bufSize := 64 * 1024 // 64KB default
	if len(bufferSize) > 0 && bufferSize[0] > 0 {
		bufSize = bufferSize[0]
	}

	// Create or truncate the output file
	file, err := os.Create(outputFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create a buffered writer for efficiency
	writer := bufio.NewWriterSize(file, bufSize)
	defer writer.Flush()

	// Write the opening bracket of the JSON array
	if _, err := writer.WriteString("["); err != nil {
		return 0, fmt.Errorf("failed to write opening bracket: %w", err)
	}

	totalCount := 0
	for _, inputPath := range inputFilePaths {
		// Open the input file
		inputFile, err := os.Open(inputPath)
		if err != nil {
			return totalCount, fmt.Errorf("failed to open input file %s: %w", inputPath, err)
		}

		// Create a JSON decoder for the input file
		decoder := json.NewDecoder(bufio.NewReaderSize(inputFile, bufSize))

		// Read the opening bracket
		t, err := decoder.Token()
		if err != nil {
			inputFile.Close()
			return totalCount, fmt.Errorf("failed to read opening bracket from %s: %w", inputPath, err)
		}
		if delim, ok := t.(json.Delim); !ok || delim != '[' {
			inputFile.Close()
			return totalCount, fmt.Errorf("expected opening bracket in %s, got %v", inputPath, t)
		}

		// Process each object in the array
		fileCount := 0
		for decoder.More() {
			// Read the next object
			var obj json.RawMessage
			if err := decoder.Decode(&obj); err != nil {
				inputFile.Close()
				return totalCount, fmt.Errorf("failed to decode object in %s: %w", inputPath, err)
			}

			// Write comma before object (except for the first object overall)
			if totalCount > 0 {
				if _, err := writer.WriteString(","); err != nil {
					inputFile.Close()
					return totalCount, fmt.Errorf("failed to write comma separator: %w", err)
				}
			}

			// Write the object
			if _, err := writer.Write(obj); err != nil {
				inputFile.Close()
				return totalCount, fmt.Errorf("failed to write object: %w", err)
			}

			fileCount++
			totalCount++

			// Periodically flush for very large files
			if totalCount%1000 == 0 {
				if err := writer.Flush(); err != nil {
					inputFile.Close()
					return totalCount, fmt.Errorf("failed to flush data: %w", err)
				}
			}
		}

		// Read the closing bracket
		t, err = decoder.Token()
		if err != nil {
			inputFile.Close()
			return totalCount, fmt.Errorf("failed to read closing bracket from %s: %w", inputPath, err)
		}
		if delim, ok := t.(json.Delim); !ok || delim != ']' {
			inputFile.Close()
			return totalCount, fmt.Errorf("expected closing bracket in %s, got %v", inputPath, t)
		}

		// Close the input file
		inputFile.Close()
	}

	// Write the closing bracket of the JSON array
	if _, err := writer.WriteString("]"); err != nil {
		return totalCount, fmt.Errorf("failed to write closing bracket: %w", err)
	}

	// Flush any buffered data to the file
	if err := writer.Flush(); err != nil {
		return totalCount, fmt.Errorf("failed to flush data to file: %w", err)
	}

	return totalCount, nil
}

// WriteObjectsToJsonArrayFile writes a slice of JavaScript objects directly to a JSON array file.
// This is a convenience function that combines ObjectsToJsonLines and WriteJsonLinesToArrayFile.
// It streams the output to minimize memory usage.
//
// Parameters:
//   - objects: An array of JavaScript objects to write to the file.
//   - outputFilePath: The path where the resulting JSON array file will be written.
//   - bufferSize: Optional buffer size in bytes (default: 64KB).
//
// Returns:
//   - The count of objects written to the file.
//   - An error if the operation failed.
//
// Example:
//
//	objects := [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]
//	count, err := streamloader.WriteObjectsToJsonArrayFile(objects, "output.json")
//	// Will write '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]' to output.json
func (s StreamLoader) WriteObjectsToJsonArrayFile(objects []interface{}, outputFilePath string, bufferSize ...int) (int, error) {
	// Set default buffer size if not provided
	bufSize := 64 * 1024 // 64KB default
	if len(bufferSize) > 0 && bufferSize[0] > 0 {
		bufSize = bufferSize[0]
	}

	// Create or truncate the output file
	file, err := os.Create(outputFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create a buffered writer for efficiency
	writer := bufio.NewWriterSize(file, bufSize)
	defer writer.Flush()

	// Write the opening bracket of the JSON array
	if _, err := writer.WriteString("["); err != nil {
		return 0, fmt.Errorf("failed to write opening bracket: %w", err)
	}

	// Process each object
	count := 0
	for i, obj := range objects {
		// Write comma separator for all but the first object
		if i > 0 {
			if _, err := writer.WriteString(","); err != nil {
				return count, fmt.Errorf("failed to write comma separator: %w", err)
			}
		}

		// Serialize the object to JSON
		objBytes, err := json.Marshal(obj)
		if err != nil {
			return count, fmt.Errorf("failed to encode object at index %d: %w", i, err)
		}

		// Write the object
		if _, err := writer.Write(objBytes); err != nil {
			return count, fmt.Errorf("failed to write object: %w", err)
		}

		count++

		// Periodically flush for very large datasets
		if count%1000 == 0 {
			if err := writer.Flush(); err != nil {
				return count, fmt.Errorf("failed to flush data: %w", err)
			}
		}
	}

	// Write the closing bracket of the JSON array
	if _, err := writer.WriteString("]"); err != nil {
		return count, fmt.Errorf("failed to write closing bracket: %w", err)
	}

	// Flush any buffered data to the file
	if err := writer.Flush(); err != nil {
		return count, fmt.Errorf("failed to flush data to file: %w", err)
	}

	return count, nil
}

// WriteCompressedObjectsToJsonArrayFile writes a slice of JavaScript objects to a JSON array file
// using compression for memory efficiency. The objects are first converted to JSONL format,
// then compressed with gzip, and finally streamed to the output file.
//
// Parameters:
//   - objects: An array of JavaScript objects to write to the file.
//   - outputFilePath: The path where the resulting JSON array file will be written.
//   - compressionLevel: Optional compression level (0-9, default is gzip.DefaultCompression).
//   - bufferSize: Optional buffer size in bytes (default: 64KB).
//
// Returns:
//   - The count of objects written to the file.
//   - An error if the operation failed.
//
// Example:
//
//	objects := [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]
//	count, err := streamloader.WriteCompressedObjectsToJsonArrayFile(objects, "output.json")
//	// Will write a JSON array with the objects to output.json, using compression for efficiency
func (s StreamLoader) WriteCompressedObjectsToJsonArrayFile(objects []interface{}, outputFilePath string, compressionLevel ...int) (int, error) {
	// Get compression level, if provided
	level := gzip.DefaultCompression
	if len(compressionLevel) > 0 && compressionLevel[0] >= gzip.NoCompression && compressionLevel[0] <= gzip.BestCompression {
		level = compressionLevel[0]
	}

	// First compress the objects to JSONL format
	compressedData, err := s.ObjectsToCompressedJsonLines(objects, level)
	if err != nil {
		return 0, fmt.Errorf("failed to compress objects: %w", err)
	}

	// Then write the compressed data to the output file as a JSON array
	return s.WriteCompressedJsonLinesToArrayFile(compressedData, outputFilePath)
}

// WriteMultipleCompressedJsonLinesToArrayFile takes multiple compressed JSON lines strings,
// decompresses them, and writes them as a single JSON array to a file. This is useful
// for combining data from multiple sources or batches while maintaining memory efficiency.
//
// Parameters:
//   - compressedJsonLinesArray: An array of base64-encoded, gzip-compressed JSONL strings.
//   - outputFilePath: The path where the resulting JSON array file will be written.
//   - bufferSize: Optional buffer size in bytes (default: 64KB).
//
// Returns:
//   - The total count of objects written to the file.
//   - An error if the operation failed.
//
// Example:
//
//	compressedBatch1 := "H4sIAAAAAAACA6tWykvMTVWyUrJSMjQ..." // Compressed JSON lines
//	compressedBatch2 := "H4sIAAAAAAAAA6tWSk4uSixJVbJSMjY..." // Another batch
//	count, err := streamloader.WriteMultipleCompressedJsonLinesToArrayFile(
//	    []string{compressedBatch1, compressedBatch2}, "combined.json")
//	// Will write a single combined JSON array to combined.json
func (StreamLoader) WriteMultipleCompressedJsonLinesToArrayFile(compressedJsonLinesArray []string, outputFilePath string, bufferSize ...int) (int, error) {
	// Set default buffer size if not provided
	bufSize := 64 * 1024 // 64KB default
	if len(bufferSize) > 0 && bufferSize[0] > 0 {
		bufSize = bufferSize[0]
	}

	// Create or truncate the output file
	file, err := os.Create(outputFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create a buffered writer for efficiency
	writer := bufio.NewWriterSize(file, bufSize)
	defer writer.Flush()

	// Write the opening bracket of the JSON array
	if _, err := writer.WriteString("["); err != nil {
		return 0, fmt.Errorf("failed to write opening bracket: %w", err)
	}

	totalCount := 0
	isFirstObject := true

	// Process each compressed JSON lines string
	for compressedIndex, compressedJsonLines := range compressedJsonLinesArray {
		if compressedJsonLines == "" {
			continue // Skip empty strings
		}

		// Decode base64 data
		compressedData, err := base64.StdEncoding.DecodeString(compressedJsonLines)
		if err != nil {
			return totalCount, fmt.Errorf("failed to decode base64 data at index %d: %w", compressedIndex, err)
		}

		// Set up the gzip reader to decompress the data
		gzReader, err := gzip.NewReader(bytes.NewReader(compressedData))
		if err != nil {
			return totalCount, fmt.Errorf("failed to create gzip reader at index %d: %w", compressedIndex, err)
		}

		// Process the decompressed JSON lines
		scanner := bufio.NewScanner(gzReader)
		// For very large lines, increase the scanner buffer size
		scanner.Buffer(make([]byte, bufSize), 10*bufSize)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue // Skip empty lines
			}

			// Write comma separator for all but the first object
			if !isFirstObject {
				if _, err := writer.WriteString(","); err != nil {
					gzReader.Close()
					return totalCount, fmt.Errorf("failed to write comma separator: %w", err)
				}
			} else {
				isFirstObject = false
			}

			// Write the JSON object to the file
			if _, err := writer.WriteString(line); err != nil {
				gzReader.Close()
				return totalCount, fmt.Errorf("failed to write JSON object: %w", err)
			}

			totalCount++
		}

		if err := scanner.Err(); err != nil {
			gzReader.Close()
			return totalCount, fmt.Errorf("error reading decompressed JSON lines at index %d: %w", compressedIndex, err)
		}

		// Close the gzip reader
		gzReader.Close()
	}

	// Write the closing bracket of the JSON array
	if _, err := writer.WriteString("]"); err != nil {
		return totalCount, fmt.Errorf("failed to write closing bracket: %w", err)
	}

	// Flush any buffered data to the file
	if err := writer.Flush(); err != nil {
		return totalCount, fmt.Errorf("failed to flush data to file: %w", err)
	}

	return totalCount, nil
}

// WriteMultipleJsonLinesToArrayFile takes multiple JSON lines strings and writes them
// as a single JSON array to a file. It streams the output to minimize memory usage.
//
// Parameters:
//   - jsonLinesArray: An array of strings containing JSONL-formatted data.
//   - outputFilePath: The path where the resulting JSON array file will be written.
//   - bufferSize: Optional buffer size in bytes (default: 64KB).
//
// Returns:
//   - The total count of objects written to the file.
//   - An error if the operation failed.
//
// Example:
//
//	batch1 := '{"id":1,"name":"Alice"}\n{"id":2,"name":"Bob"}'
//	batch2 := '{"id":3,"name":"Charlie"}\n{"id":4,"name":"Dave"}'
//	count, err := streamloader.WriteMultipleJsonLinesToArrayFile(
//	    []string{batch1, batch2}, "combined.json")
//	// Will write '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"},{"id":3,"name":"Charlie"},{"id":4,"name":"Dave"}]'
//	// to combined.json
func (StreamLoader) WriteMultipleJsonLinesToArrayFile(jsonLinesArray []string, outputFilePath string, bufferSize ...int) (int, error) {
	// Set default buffer size if not provided
	bufSize := 64 * 1024 // 64KB default
	if len(bufferSize) > 0 && bufferSize[0] > 0 {
		bufSize = bufferSize[0]
	}

	// Create or truncate the output file
	file, err := os.Create(outputFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create a buffered writer for efficiency
	writer := bufio.NewWriterSize(file, bufSize)
	defer writer.Flush()

	// Write the opening bracket of the JSON array
	if _, err := writer.WriteString("["); err != nil {
		return 0, fmt.Errorf("failed to write opening bracket: %w", err)
	}

	totalCount := 0
	isFirstObject := true

	// Process each JSON lines string
	for batchIndex, jsonLines := range jsonLinesArray {
		if jsonLines == "" {
			continue // Skip empty strings
		}

		// Process the JSON lines
		scanner := bufio.NewScanner(strings.NewReader(jsonLines))
		// For very large lines, increase the scanner buffer size
		scanner.Buffer(make([]byte, bufSize), 10*bufSize)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue // Skip empty lines
			}

			// Write comma separator for all but the first object
			if !isFirstObject {
				if _, err := writer.WriteString(","); err != nil {
					return totalCount, fmt.Errorf("failed to write comma separator: %w", err)
				}
			} else {
				isFirstObject = false
			}

			// Validate that the line is a valid JSON object
			var obj interface{}
			if err := json.Unmarshal([]byte(line), &obj); err != nil {
				return totalCount, fmt.Errorf("invalid JSON at batch %d: %w", batchIndex, err)
			}
			
			// Write the JSON object to the file
			if _, err := writer.WriteString(line); err != nil {
				return totalCount, fmt.Errorf("failed to write JSON object from batch %d: %w", batchIndex, err)
			}

			totalCount++
		}

		if err := scanner.Err(); err != nil {
			return totalCount, fmt.Errorf("error reading JSON lines from batch %d: %w", batchIndex, err)
		}
	}

	// Write the closing bracket of the JSON array
	if _, err := writer.WriteString("]"); err != nil {
		return totalCount, fmt.Errorf("failed to write closing bracket: %w", err)
	}

	// Flush any buffered data to the file
	if err := writer.Flush(); err != nil {
		return totalCount, fmt.Errorf("failed to flush data to file: %w", err)
	}

	return totalCount, nil
}

// JsonLinesToObjects takes a JSONL-formatted string and converts it to a slice of objects.
// Each line in the input is parsed as a separate JSON object.
//
// Parameters:
//   - jsonLines: A string containing JSONL-formatted data, with one JSON object per line.
//
// Returns:
//   - A slice of parsed objects ([]interface{}).
//   - An error if any line contains invalid JSON.
//
// Example:
//
//     jsonLines := `{"id":1,"name":"Alice"}
//     {"id":2,"name":"Bob"}`
//     objects, err := streamloader.JsonLinesToObjects(jsonLines)
//     // objects will be [{id:1, name:"Alice"}, {id:2, name:"Bob"}]
func (StreamLoader) JsonLinesToObjects(jsonLines string) ([]interface{}, error) {
	if jsonLines == "" {
		return []interface{}{}, nil
	}

	var objects []interface{}
	scanner := bufio.NewScanner(strings.NewReader(jsonLines))
	
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip empty lines
		}

		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			return nil, fmt.Errorf("invalid JSON at line %d: %w", lineNum, err)
		}
		objects = append(objects, obj)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading JSON lines: %w", err)
	}

	return objects, nil
}

// CompressedJsonLinesToObjects takes a base64-encoded, gzip-compressed JSONL string
// and converts it to a slice of objects. It decodes the base64 data, decompresses it,
// and parses each line as a separate JSON object.
//
// Parameters:
//   - compressedJsonLines: A base64-encoded string containing gzip-compressed JSONL data.
//
// Returns:
//   - A slice of parsed objects ([]interface{}).
//   - An error if decompression fails or any line contains invalid JSON.
//
// Example:
//
//     compressedData := "H4sIAAAAAAAA/6tWSk5OLCpKVbJSMjA2M9RRKsgsVrIyBHITKzNSixQUQPLJ..."
//     objects, err := streamloader.CompressedJsonLinesToObjects(compressedData)
//     // objects will be the decompressed and parsed objects
func (s StreamLoader) CompressedJsonLinesToObjects(compressedJsonLines string) ([]interface{}, error) {
	// Decode base64 data
	compressedData, err := base64.StdEncoding.DecodeString(compressedJsonLines)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Set up the gzip reader to decompress the data
	gzReader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Read all decompressed data
	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// Convert to string and parse using JsonLinesToObjects
	return (&StreamLoader{}).JsonLinesToObjects(string(decompressed))
}

func init() {
	modules.Register("k6/x/streamloader", new(StreamLoader))
}
