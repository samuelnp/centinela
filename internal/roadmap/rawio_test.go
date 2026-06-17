package roadmap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// minimalRoadmapJSON is the minimal valid roadmap with a real phase.
const minimalRoadmapJSON = `{"phases":[{"name":"Phase 0","features":[{"name":"f1"}]}]}`

// TestReadRawRoadmap_HappyPath parses a valid roadmap.json.
func TestReadRawRoadmap_HappyPath(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(minimalRoadmapJSON), 0644) //nolint:errcheck
	doc, err := readRawRoadmap(p)
	if err != nil {
		t.Fatalf("readRawRoadmap: %v", err)
	}
	if len(doc.phases) != 1 {
		t.Errorf("expected 1 phase, got %d", len(doc.phases))
	}
}

// TestReadRawRoadmap_Missing returns an error for a missing file.
func TestReadRawRoadmap_Missing(t *testing.T) {
	if _, err := readRawRoadmap("/nonexistent/roadmap.json"); err == nil {
		t.Error("expected error for missing file")
	}
}

// TestReadRawRoadmap_Corrupt returns an error for invalid JSON.
func TestReadRawRoadmap_Corrupt(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte("{bad"), 0644) //nolint:errcheck
	if _, err := readRawRoadmap(p); err == nil {
		t.Error("expected error for corrupt JSON")
	}
}

// TestReadRawRoadmap_InvalidPhases returns an error when phases is not an array.
func TestReadRawRoadmap_InvalidPhases(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(`{"phases":"not-an-array"}`), 0644) //nolint:errcheck
	if _, err := readRawRoadmap(p); err == nil {
		t.Error("expected error for phases not being an array")
	}
}

// TestWriteAtomic_Success writes and reads back a file.
func TestWriteAtomic_Success(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "out.json")
	if err := writeAtomic(p, []byte(`{"x":1}`)); err != nil {
		t.Fatalf("writeAtomic: %v", err)
	}
	got, _ := os.ReadFile(p)
	if string(got) != `{"x":1}` {
		t.Errorf("unexpected content: %s", got)
	}
}

// TestCompactBytes_Valid encodes a struct to a single-line JSON.
func TestCompactBytes_Valid(t *testing.T) {
	raw, err := compactBytes(Feature{Name: "my-slug"})
	if err != nil {
		t.Fatalf("compactBytes: %v", err)
	}
	s := string(raw)
	if strings.Contains(s, "\n") {
		t.Error("compactBytes must not contain newlines")
	}
	if !strings.Contains(s, `"my-slug"`) {
		t.Errorf("unexpected compactBytes output: %s", s)
	}
}

// TestWriteRawRoadmap_RoundTrip writes and verifies format.
func TestWriteRawRoadmap_RoundTrip(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(minimalRoadmapJSON), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	if err := writeRawRoadmap(p, doc); err != nil {
		t.Fatalf("writeRawRoadmap: %v", err)
	}
	got, _ := os.ReadFile(p)
	if !strings.Contains(string(got), "Phase 0") {
		t.Errorf("phase name missing after round-trip: %s", got)
	}
}
