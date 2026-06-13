package roadmap

import (
	"os"
	"path/filepath"
	"testing"
)

// writeAtomic must surface MkdirAll failures when a path component is a file.
func TestWriteAtomic_MkdirAllError(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	// Parent dir is a regular file, so MkdirAll must fail.
	if err := writeAtomic(filepath.Join(blocker, "child", "roadmap.json"), []byte("{}")); err == nil {
		t.Fatal("expected MkdirAll error, got nil")
	}
}

// compactBytes must propagate encoder errors for unmarshalable values.
func TestCompactBytes_EncodeError(t *testing.T) {
	if _, err := compactBytes(make(chan int)); err == nil {
		t.Fatal("expected encode error for channel, got nil")
	}
}

// readRawRoadmap surfaces both missing-file and malformed-JSON errors.
func TestReadRawRoadmap_Errors(t *testing.T) {
	if _, err := readRawRoadmap(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Fatal("expected error for missing file")
	}
	bad := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(bad, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := readRawRoadmap(bad); err == nil {
		t.Fatal("expected error for malformed json")
	}
}

// readRawRoadmap rejects a valid top-level object with a non-array phases key.
func TestReadRawRoadmap_BadPhases(t *testing.T) {
	p := filepath.Join(t.TempDir(), "r.json")
	if err := os.WriteFile(p, []byte(`{"phases": 42}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := readRawRoadmap(p); err == nil {
		t.Fatal("expected error for non-array phases")
	}
}

// BootstrapFeatures returns nil for a nil roadmap and the names otherwise.
func TestBootstrapFeatures_NilAndPopulated(t *testing.T) {
	if got := BootstrapFeatures(nil); got != nil {
		t.Fatalf("expected nil for nil roadmap, got %v", got)
	}
	r := &Roadmap{Phases: []Phase{
		{Name: "Phase 0: Bootstrap", Features: []Feature{{Name: "a"}, {Name: "b"}}},
		{Name: "Phase 1: Other", Features: []Feature{{Name: "c"}}},
	}}
	got := BootstrapFeatures(r)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected bootstrap features: %v", got)
	}
}
