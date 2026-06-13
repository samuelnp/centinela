package roadmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestDecodePhase_InvalidJSON returns an error.
func TestDecodePhase_InvalidJSON(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(`{"phases":[{bad}]}`), 0644) //nolint:errcheck
	if _, err := readRawRoadmap(p); err == nil {
		t.Error("expected error for invalid phase JSON")
	}
}

// TestSetPhase_ReflectsInDirtyMap marks the phase dirty.
func TestSetPhase_ReflectsInDirtyMap(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(minimalRoadmapJSON), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	ph, _ := doc.decodePhase(0)
	ph.Features = append(ph.Features, json.RawMessage(`{"name":"extra"}`))
	if err := doc.setPhase(0, ph); err != nil {
		t.Fatalf("setPhase: %v", err)
	}
	if _, ok := doc.dirty[0]; !ok {
		t.Error("phase 0 must be marked dirty")
	}
}

// TestFeatureName_Invalid returns error for corrupt JSON.
func TestFeatureName_Invalid(t *testing.T) {
	if _, err := featureName(json.RawMessage(`{bad`)); err == nil {
		t.Error("expected error for corrupt feature JSON")
	}
}

// TestPhaseFeatureNames_CorruptFeature returns error.
func TestPhaseFeatureNames_CorruptFeature(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	// Phase has a feature that is not valid JSON for name extraction
	os.WriteFile(p, []byte(`{"phases":[{"name":"P0","features":[{bad]}]}`), 0644) //nolint:errcheck
	// readRawRoadmap will fail parsing the phases array
	if _, err := readRawRoadmap(p); err == nil {
		t.Error("expected error for corrupt phases array")
	}
}
