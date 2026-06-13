package roadmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPhaseFeatureNames_HappyPath returns a name→phase map.
func TestPhaseFeatureNames_HappyPath(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(`{"phases":[{"name":"P0","features":[{"name":"f1"}]},{"name":"P1","features":[{"name":"f2"}]}]}`), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	names, err := doc.phaseFeatureNames()
	if err != nil {
		t.Fatalf("phaseFeatureNames: %v", err)
	}
	if names["f1"] != "P0" || names["f2"] != "P1" {
		t.Errorf("unexpected names: %v", names)
	}
}

// TestAppendBacklog_NewPhase creates the Backlog phase when absent.
func TestAppendBacklog_NewPhase(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(minimalRoadmapJSON), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	entry, _ := compactBytes(Feature{Name: "new-finding"})
	if err := doc.appendBacklog(entry); err != nil {
		t.Fatalf("appendBacklog: %v", err)
	}
	writeRawRoadmap(p, doc) //nolint:errcheck
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), "Backlog") {
		t.Error("Backlog phase must appear after appendBacklog")
	}
	if !strings.Contains(string(data), "new-finding") {
		t.Error("entry must appear in Backlog")
	}
}

// TestAppendBacklog_ExistingPhase appends to an existing Backlog.
func TestAppendBacklog_ExistingPhase(t *testing.T) {
	src := `{"phases":[{"name":"Phase 0","features":[]},{"name":"Backlog","features":[{"name":"prior"}]}]}`
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(src), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	entry, _ := compactBytes(Feature{Name: "new-entry"})
	doc.appendBacklog(entry) //nolint:errcheck
	writeRawRoadmap(p, doc)  //nolint:errcheck
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), "prior") || !strings.Contains(string(data), "new-entry") {
		t.Errorf("both entries must survive: %s", data)
	}
}

// TestFeatureName_Valid extracts name from raw feature JSON.
func TestFeatureName_Valid(t *testing.T) {
	raw := json.RawMessage(`{"name":"my-feat","dependsOn":[]}`)
	got, err := featureName(raw)
	if err != nil || got != "my-feat" {
		t.Errorf("featureName: %v, %v", got, err)
	}
}

// TestEncodePhase produces valid JSON with name and features.
func TestEncodePhase(t *testing.T) {
	p := &rawPhase{Name: "P0", Features: []json.RawMessage{json.RawMessage(`{"name":"x"}`)}}
	raw, err := encodePhase(p)
	if err != nil {
		t.Fatalf("encodePhase: %v", err)
	}
	if !strings.Contains(string(raw), "P0") {
		t.Errorf("phase name missing: %s", raw)
	}
}
