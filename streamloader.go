// streamloader.go
package streamloader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"go.k6.io/k6/js/modules"
)

type TrafficRecordingSample struct {
	Method     string            `json:"method"`
	RequestURI string            `json:"requestURI"`
	Headers    map[string]string `json:"headers"`
	Content    string            `json:"content"`
}

// StreamLoader is the k6/x/streamloader module
// It provides a method to load all samples from a JSON array file
// using a small buffer to minimize intermediate memory spikes.
type StreamLoader struct{}

// LoadSamples opens the given file, streams and parses its
// JSON array of TrafficRecordingSample, and returns the slice.
func (StreamLoader) LoadSamples(filePath string) ([]TrafficRecordingSample, error) {
	// 1) Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 2) Wrap in a buffered reader (64 KB)
	reader := bufio.NewReaderSize(file, 64*1024)

	// 3) Create JSON decoder
	dec := json.NewDecoder(reader)

	// 4) Read opening '[' token
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return nil, fmt.Errorf("expected JSON array, got %v", tok)
	}

	// 5) Decode each element into the slice
	var samples []TrafficRecordingSample
	for dec.More() {
		var s TrafficRecordingSample
		if err := dec.Decode(&s); err != nil {
			return nil, err
		}
		samples = append(samples, s)
	}

	// 6) Consume closing ']' token
	if _, err := dec.Token(); err != nil {
		return nil, err
	}

	return samples, nil
}

func init() {
	modules.Register("k6/x/streamloader", new(StreamLoader))
}
