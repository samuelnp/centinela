package roadmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRender_UntouchedPhase produces valid JSON with the original phase name.
func TestRender_UntouchedPhase(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	src := `{"phases":[{"name":"Phase 0","features":[{"name":"f1"}]}]}`
	os.WriteFile(p, []byte(src), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	data, err := doc.render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(string(data), "Phase 0") || !strings.Contains(string(data), "f1") {
		t.Errorf("untouched phase fields missing: %s", data)
	}
}

// TestRender_ExtraTopLevelKeySorted writes sorted non-phases keys.
func TestRender_ExtraTopLevelKeySorted(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	src := `{"phases":[{"name":"P0","features":[]}],"zzz":"last","aaa":"first"}`
	os.WriteFile(p, []byte(src), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	data, err := doc.render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	s := string(data)
	zzz := strings.Index(s, "zzz")
	aaa := strings.Index(s, "aaa")
	if aaa == -1 || zzz == -1 || aaa > zzz {
		t.Errorf("sorted keys: aaa must appear before zzz in: %s", s)
	}
}

// TestIndentValue returns a re-indented JSON string.
func TestIndentValue(t *testing.T) {
	got, err := indentValue(json.RawMessage(`{"a":1,"b":2}`), "  ")
	if err != nil {
		t.Fatalf("indentValue: %v", err)
	}
	if !strings.Contains(got, `"a"`) {
		t.Errorf("indentValue should preserve keys: %q", got)
	}
}

// TestIndentValue_Invalid returns an error for non-JSON.
func TestIndentValue_Invalid(t *testing.T) {
	if _, err := indentValue(json.RawMessage(`{bad`), "  "); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// TestBacklogPhaseIndex_Present returns the correct index.
func TestBacklogPhaseIndex_Present(t *testing.T) {
	src := `{"phases":[{"name":"P0","features":[]},{"name":"Backlog","features":[]}]}`
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(src), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	idx, err := doc.backlogPhaseIndex()
	if err != nil || idx != 1 {
		t.Errorf("expected idx=1, got %d, err=%v", idx, err)
	}
}

// TestBacklogPhaseIndex_Absent returns -1.
func TestBacklogPhaseIndex_Absent(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(minimalRoadmapJSON), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	idx, err := doc.backlogPhaseIndex()
	if err != nil || idx != -1 {
		t.Errorf("expected idx=-1, got %d, err=%v", idx, err)
	}
}
