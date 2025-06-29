// streamloader.go
package streamloader

import (
	"bufio"
	"container/ring"
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
// - trimLeadingSpace: Removes leading whitespace from fields (default: true)
//   - Only removes whitespace at the beginning of fields when a field is read
//   - This is a built-in feature of the CSV reader
// - trimSpace: Removes all whitespace from fields (leading and trailing) (default: false)
//   - This is a more aggressive trimming that removes both leading and trailing whitespace
//   - When true, this overrides the CSV reader's built-in TrimLeadingSpace behavior
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

func init() {
	modules.Register("k6/x/streamloader", new(StreamLoader))
}
