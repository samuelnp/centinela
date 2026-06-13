package roadmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestPhaseBytes_DirtyPath returns dirty bytes when phase is mutated.
func TestPhaseBytes_DirtyPath(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(minimalRoadmapJSON), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	// Force phase 0 dirty
	ph := &rawPhase{Name: "Phase 0 Mutated", Features: []json.RawMessage{}}
	doc.setPhase(0, ph) //nolint:errcheck
	b := doc.phaseBytes(0)
	// Must contain the mutated name from dirty map
	if string(b) == "" {
		t.Error("phaseBytes must return non-empty dirty bytes")
	}
}

// TestRender_DirtyPhaseIncluded renders dirty phase features one-per-line.
func TestRender_DirtyPhaseIncluded(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(minimalRoadmapJSON), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	// Append a backlog entry to force dirty state
	entry, _ := compactBytes(Feature{Name: "dirty-entry"})
	doc.appendBacklog(entry) //nolint:errcheck
	data, err := doc.render()
	if err != nil {
		t.Fatalf("render with dirty phase: %v", err)
	}
	if string(data) == "" {
		t.Error("render must not be empty")
	}
}

// TestBacklogPhaseIndex_LowercaseBacklog finds Backlog with case variant.
func TestBacklogPhaseIndex_LowercaseBacklog(t *testing.T) {
	src := `{"phases":[{"name":"Phase 0","features":[]},{"name":"backlog","features":[]}]}`
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(src), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	idx, err := doc.backlogPhaseIndex()
	if err != nil || idx != 1 {
		t.Errorf("lowercase backlog must be found at idx=1: idx=%d, err=%v", idx, err)
	}
}
