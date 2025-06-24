// streamloader.go
package streamloader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.k6.io/k6/js/modules"
)

// StreamLoader is the k6/x/streamloader module.
// It provides LoadJSON for reading large JSON files efficiently
// using a small buffer and supporting standard JSON arrays, NDJSON, or JSON objects.
type StreamLoader struct{}

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
