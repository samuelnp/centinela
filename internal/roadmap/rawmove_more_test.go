package roadmap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestKnownPhaseList_ListsNonBacklogPhases returns comma-joined non-Backlog names.
func TestKnownPhaseList_ListsNonBacklogPhases(t *testing.T) {
	src := `{"phases":[{"name":"Phase 0","features":[]},{"name":"Phase 5","features":[]},{"name":"Backlog","features":[]}]}`
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(src), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	list := doc.knownPhaseList()
	if !strings.Contains(list, "Phase 0") || !strings.Contains(list, "Phase 5") {
		t.Errorf("known phase list missing phases: %s", list)
	}
	if strings.Contains(list, "Backlog") {
		t.Errorf("Backlog must not appear in known phase list: %s", list)
	}
}

// TestRemoveBacklogFeature_NoMatchIsIdempotent removes nothing when slug absent.
func TestRemoveBacklogFeature_NoMatchIsIdempotent(t *testing.T) {
	src := `{"phases":[{"name":"Backlog","features":[{"name":"keep-me"}]}]}`
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(src), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	// remove a slug that is NOT present — must not corrupt the phase
	doc.removeBacklogFeature(0, "nonexistent") //nolint:errcheck
	writeRawRoadmap(p, doc)                    //nolint:errcheck
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), "keep-me") {
		t.Error("keep-me must still be present after no-op removal")
	}
}

// TestPhaseName_InvalidJSON returns error.
func TestPhaseName_InvalidJSON(t *testing.T) {
	if _, err := phaseName([]byte(`{bad`)); err == nil {
		t.Error("expected error for corrupt phase JSON")
	}
}
