package streamloader

import (
	"os"
	"testing"
)

func TestLoadSamples(t *testing.T) {
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
	if samples[0].Method != "GET" || samples[0].RequestURI != "/foo" || samples[0].Headers["A"] != "B" || samples[0].Content != "abc" {
		t.Errorf("unexpected first sample: %+v", samples[0])
	}
	if samples[1].Method != "POST" || samples[1].RequestURI != "/bar" || samples[1].Headers["C"] != "D" || samples[1].Content != "def" {
		t.Errorf("unexpected second sample: %+v", samples[1])
	}
}
