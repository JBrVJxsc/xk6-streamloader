package streamloader

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestLoadJSON_ArrayFormat(t *testing.T) {
	jsonData := `[
	  {"method": "GET", "requestURI": "/foo", "headers": {"A": "B"}, "content": "abc"},
	  {"method": "POST", "requestURI": "/bar", "headers": {"C": "D"}, "content": "def"}
	]`

	tmpfile, err := os.CreateTemp("", "objects-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	loader := StreamLoader{}
	objects, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}

	if len(objects) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(objects))
	}

	// First object
	o0 := objects[0]
	if m, ok := o0["method"].(string); !ok || m != "GET" {
		t.Errorf("expected method GET, got %v", o0["method"])
	}
	if uri, ok := o0["requestURI"].(string); !ok || uri != "/foo" {
		t.Errorf("expected requestURI /foo, got %v", o0["requestURI"])
	}
	h, ok := o0["headers"].(map[string]interface{})
	if !ok || h["A"].(string) != "B" {
		t.Errorf("expected header A:B, got %v", o0["headers"])
	}
	if c, ok := o0["content"].(string); !ok || c != "abc" {
		t.Errorf("expected content abc, got %v", o0["content"])
	}

	// Second object
	o1 := objects[1]
	if m, ok := o1["method"].(string); !ok || m != "POST" {
		t.Errorf("expected method POST, got %v", o1["method"])
	}
	if uri, ok := o1["requestURI"].(string); !ok || uri != "/bar" {
		t.Errorf("expected requestURI /bar, got %v", o1["requestURI"])
	}
	h2, ok := o1["headers"].(map[string]interface{})
	if !ok || h2["C"].(string) != "D" {
		t.Errorf("expected header C:D, got %v", o1["headers"])
	}
	if c, ok := o1["content"].(string); !ok || c != "def" {
		t.Errorf("expected content def, got %v", o1["content"])
	}
}

func TestLoadJSON_EmptyArray(t *testing.T) {
	jsonData := `[]`
	tmpfile, err := os.CreateTemp("", "objects-empty-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	objects, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	if len(objects) != 0 {
		t.Errorf("expected 0 objects, got %d", len(objects))
	}
}

func TestLoadJSON_InvalidJSON(t *testing.T) {
	jsonData := `{not valid json`
	tmpfile, err := os.CreateTemp("", "objects-bad-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	_, err = loader.LoadJSON(tmpfile.Name())
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoadJSON_MissingFile(t *testing.T) {
	loader := StreamLoader{}
	_, err := loader.LoadJSON("no_such_file.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadJSON_LargeArray(t *testing.T) {
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

	tmpfile, err := os.CreateTemp("", "objects-large-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(b.String())); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	objects, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	if len(objects) != 1000 {
		t.Errorf("expected 1000 objects, got %d", len(objects))
	}
	if uri, ok := objects[0]["requestURI"].(string); !ok || uri != "/bulk/0" {
		t.Errorf("unexpected first object URI: %v", objects[0]["requestURI"])
	}
	if uri, ok := objects[999]["requestURI"].(string); !ok || uri != "/bulk/999" {
		t.Errorf("unexpected last object URI: %v", objects[999]["requestURI"])
	}
}

func TestLoadJSON_WrongFieldName(t *testing.T) {
	jsonData := `[{"method": "GET", "request_uri": "/foo", "headers": {"A": "B"}, "content": "abc"}]`
	tmpfile, err := os.CreateTemp("", "objects-wrongfield-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpfile.Close()
	loader := StreamLoader{}
	objects, err := loader.LoadJSON(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadJSON failed: %v", err)
	}
	if _, ok := objects[0]["requestURI"]; ok {
		t.Errorf("did not expect requestURI key, but found %v", objects[0]["requestURI"])
	}
	if v, ok := objects[0]["request_uri"]; !ok || v != "/foo" {
		t.Errorf("expected request_uri key with value /foo, got %v", v)
	}
}
