package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestRenderPromoteEvaluatorContext_WithSource includes all required fields.
func TestRenderPromoteEvaluatorContext_WithSource(t *testing.T) {
	f := &roadmap.BacklogFinding{
		Name:    "hook-timeout-config",
		Summary: "Prewrite hook timeout is hardcoded",
		Source:  &roadmap.Source{Feature: "deferred-findings-roadmap-capture", Role: "senior-engineer"},
	}
	got := RenderPromoteEvaluatorContext(f, "Phase 5 — Operability & DX")
	if !strings.Contains(got, "hook-timeout-config") {
		t.Error("finding name missing")
	}
	if !strings.Contains(got, "Prewrite hook timeout") {
		t.Error("summary missing")
	}
	if !strings.Contains(got, "Phase 5") {
		t.Error("target phase missing")
	}
	if !strings.Contains(got, "9") {
		t.Error("quality threshold missing")
	}
	if !strings.Contains(got, "acceptanceCriteria") {
		t.Error("scoring dimensions missing")
	}
	if !strings.Contains(got, "--scores") {
		t.Error("re-invocation line missing")
	}
}

// TestRenderPromoteEvaluatorContext_NoSource renders "(none)" for source.
func TestRenderPromoteEvaluatorContext_NoSource(t *testing.T) {
	f := &roadmap.BacklogFinding{
		Name:    "no-source-slug",
		Summary: "Root-level finding",
		Source:  nil,
	}
	got := RenderPromoteEvaluatorContext(f, "Phase 0: Bootstrap")
	if !strings.Contains(got, "(none)") {
		t.Errorf("(none) must appear for nil source: %s", got)
	}
}

// TestRenderPromoteEvaluatorContext_AllDimensions verifies six dimensions.
func TestRenderPromoteEvaluatorContext_AllDimensions(t *testing.T) {
	f := &roadmap.BacklogFinding{Name: "x", Summary: "s"}
	got := RenderPromoteEvaluatorContext(f, "Phase 1")
	for _, dim := range []string{"acceptanceCriteria", "userValue", "definitionClarity",
		"dependencies", "effortEstimation", "overall"} {
		if !strings.Contains(got, dim) {
			t.Errorf("dimension %q missing from evaluator context", dim)
		}
	}
}
