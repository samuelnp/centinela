package roadmap

import (
	"os"
	"strings"
	"testing"
)

// Load surfaces a ValidateDependencies error when a feature depends on an
// unknown feature.
func TestLoad_DependencyValidationError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0755); err != nil {
		t.Fatal(err)
	}
	body := `{"phases":[{"name":"P0","features":[{"name":"a","dependsOn":["ghost"]}]}]}`
	if err := os.WriteFile(RoadmapFile, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "unknown feature") {
		t.Fatalf("expected dependency validation error, got %v", err)
	}
}

// Load surfaces an unmarshal error for malformed roadmap.json.
func TestLoad_UnmarshalError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(RoadmapFile, []byte("{bad"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(); err == nil {
		t.Fatal("expected unmarshal error for malformed roadmap.json")
	}
}
