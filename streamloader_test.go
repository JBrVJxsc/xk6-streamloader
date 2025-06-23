package streamloader

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestLoadSamples_ArrayFormat(t *testing.T) {
	jsonData := `[
	  {"method": "GET", "requestURI": "/foo", "headers": {"A": "B"}, "content": "abc"},
	  {"method": "POST", "requestURI": "/bar", "headers": {"C": "D"}, "content": "def"}
	]`

	tmpfile, err := os.CreateTemp("", "samples-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	loader := StreamLoader{}
	samples, err := loader.LoadSamples(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadSamples failed: %v", err)
	}

	if len(samples) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(samples))
	}

	// First sample
	s0 := samples[0]
	if m, ok := s0["method"].(string); !ok || m != "GET" {
		t.Errorf("expected method GET, got %v", s0["method"])
	}
	if uri, ok := s0["requestURI"].(string); !ok || uri != "/foo" {
		t.Errorf("expected requestURI /foo, got %v", s0["requestURI"])
	}
	h, ok := s0["headers"].(map[string]interface{})
	if !ok || h["A"].(string) != "B" {
		t.Errorf("expected header A:B, got %v", s0["headers"])
	}
	if c, ok := s0["content"].(string); !ok || c != "abc" {
		t.Errorf("expected content abc, got %v", s0["content"])
	}

	// Second sample
	s1 := samples[1]
	if m, ok := s1["method"].(string); !ok || m != "POST" {
		t.Errorf("expected method POST, got %v", s1["method"])
	}
	if uri, ok := s1["requestURI"].(string); !ok || uri != "/bar" {
		t.Errorf("expected requestURI /bar, got %v", s1["requestURI"])
	}
	h2, ok := s1["headers"].(map[string]interface{})
	if !ok || h2["C"].(string) != "D" {
		t.Errorf("expected header C:D, got %v", s1["headers"])
	}
	if c, ok := s1["content"].(string); !ok || c != "def" {
		t.Errorf("expected content def, got %v", s1["content"])
	}
}

func TestLoadSamples_EmptyArray(t *testing.T) {
	jsonData := `[]`
	tmpfile, err := os.CreateTemp("", "samples-empty-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	samples, err := loader.LoadSamples(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadSamples failed: %v", err)
	}
	if len(samples) != 0 {
		t.Errorf("expected 0 samples, got %d", len(samples))
	}
}

func TestLoadSamples_InvalidJSON(t *testing.T) {
	jsonData := `{not valid json`
	tmpfile, err := os.CreateTemp("", "samples-bad-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	_, err = loader.LoadSamples(tmpfile.Name())
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoadSamples_MissingFile(t *testing.T) {
	loader := StreamLoader{}
	_, err := loader.LoadSamples("no_such_file.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadSamples_LargeArray(t *testing.T) {
	// Generate large JSON
	var b strings.Builder
	b.WriteString("[")
	for i := 0; i < 1000; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"method":"GET","requestURI":"/bulk/` + strconv.Itoa(i) + `","headers":{"X":"` + strconv.Itoa(i) + `"},"content":"` + strconv.Itoa(i) + `"}`)
	}
	b.WriteString("]")

	tmpfile, err := os.CreateTemp("", "samples-large-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(b.String())); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	samples, err := loader.LoadSamples(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadSamples failed: %v", err)
	}
	if len(samples) != 1000 {
		t.Errorf("expected 1000 samples, got %d", len(samples))
	}
	if uri, ok := samples[0]["requestURI"].(string); !ok || uri != "/bulk/0" {
		t.Errorf("unexpected first sample URI: %v", samples[0]["requestURI"])
	}
	if uri, ok := samples[999]["requestURI"].(string); !ok || uri != "/bulk/999" {
		t.Errorf("unexpected last sample URI: %v", samples[999]["requestURI"])
	}
}

func TestLoadSamples_WrongFieldName(t *testing.T) {
	jsonData := `[{"method": "GET", "request_uri": "/foo", "headers": {"A": "B"}, "content": "abc"}]`
	tmpfile, err := os.CreateTemp("", "samples-wrongfield-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	samples, err := loader.LoadSamples(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadSamples failed: %v", err)
	}
	if _, ok := samples[0]["requestURI"]; ok {
		t.Errorf("did not expect requestURI key, but found %v", samples[0]["requestURI"])
	}
	if v, ok := samples[0]["request_uri"]; !ok || v != "/foo" {
		t.Errorf("expected request_uri key with value /foo, got %v", v)
	}
}
