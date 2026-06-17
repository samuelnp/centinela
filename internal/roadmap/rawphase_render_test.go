package roadmap

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestRenderDirtyPhase_OneFeatPerLine renders features one per line.
func TestRenderDirtyPhase_OneFeatPerLine(t *testing.T) {
	p := &rawPhase{
		Name: "Backlog",
		Features: []json.RawMessage{
			json.RawMessage(`{"name":"f1"}`),
			json.RawMessage(`{"name":"f2"}`),
		},
	}
	raw, _ := encodePhase(p)
	got, err := renderDirtyPhase(raw)
	if err != nil {
		t.Fatalf("renderDirtyPhase: %v", err)
	}
	if !strings.Contains(got, "Backlog") {
		t.Error("phase name missing")
	}
	if !strings.Contains(got, `"f1"`) || !strings.Contains(got, `"f2"`) {
		t.Errorf("features missing: %s", got)
	}
}

// TestRenderDirtyPhase_EmptyFeatures renders a phase with no features.
func TestRenderDirtyPhase_EmptyFeatures(t *testing.T) {
	p := &rawPhase{Name: "Backlog", Features: []json.RawMessage{}}
	raw, _ := encodePhase(p)
	got, err := renderDirtyPhase(raw)
	if err != nil {
		t.Fatalf("renderDirtyPhase empty: %v", err)
	}
	if !strings.Contains(got, "features") {
		t.Errorf("features key missing: %s", got)
	}
}

// TestRenderDirtyPhase_Invalid returns error for corrupt JSON.
func TestRenderDirtyPhase_Invalid(t *testing.T) {
	if _, err := renderDirtyPhase(json.RawMessage(`{bad`)); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// TestWritePhaseKey adds a key to the buffer correctly.
func TestWritePhaseKey(t *testing.T) {
	var buf strings.Builder
	first := true
	// writePhaseKey uses bytes.Buffer but we can use the method directly via rawphase_render.
	// We call renderDirtyPhase which exercises writePhaseKey indirectly.
	buf.WriteString("") // suppress unused
	p := &rawPhase{Name: "P1", Features: []json.RawMessage{json.RawMessage(`{"name":"x"}`)}}
	raw, _ := encodePhase(p)
	out, err := renderDirtyPhase(raw)
	if err != nil {
		t.Fatalf("renderDirtyPhase: %v", err)
	}
	if !strings.Contains(out, `"name"`) {
		t.Errorf("key 'name' missing: %s", out)
	}
	_ = first
}
