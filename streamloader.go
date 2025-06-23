// streamloader.go
package streamloader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.k6.io/k6/js/modules"
)

// StreamLoader is the k6/x/streamloader module.
// It provides LoadSamples for reading large JSON files efficiently
// using a small buffer and supporting standard JSON arrays or NDJSON.
type StreamLoader struct{}

// LoadSamples opens the given file, streams and parses its JSON content into a slice of generic maps.
// By returning map[string]interface{}, we preserve the original JSON key names exactly as-is.
func (StreamLoader) LoadSamples(filePath string) ([]map[string]any, error) {
	// 1) Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 2) Buffered reader (64 KB)
	reader := bufio.NewReaderSize(file, 64*1024)

	// 3) Peek first non-whitespace byte to detect format
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

	var samples []map[string]any

	if firstByte == '[' {
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

		// Decode each object in the array into a generic map
		for dec.More() {
			var item map[string]any
			if err := dec.Decode(&item); err != nil {
				return nil, err
			}
			samples = append(samples, item)
		}

		// Consume closing ']'
		if _, err := dec.Token(); err != nil {
			return nil, err
		}
	} else {
		// Newline-delimited JSON (NDJSON) format
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var item map[string]any
			if err := json.Unmarshal([]byte(line), &item); err != nil {
				return nil, err
			}
			samples = append(samples, item)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return samples, nil
}

// isWhitespace checks for JSON whitespace characters
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t'
}

func init() {
	modules.Register("k6/x/streamloader", new(StreamLoader))
}
