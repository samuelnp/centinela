package ui

import (
	"strings"
	"testing"
)

// TestRenderRoadmapCheckpoint covers the pure roadmap-checkpoint panel renderer.
func TestRenderRoadmapCheckpoint(t *testing.T) {
	out := RenderRoadmapCheckpoint("bootstrap-feature")
	if !strings.Contains(out, "bootstrap-feature") {
		t.Fatalf("expected feature name in panel, got: %q", out)
	}
	if !strings.Contains(out, "centinela start bootstrap-feature") {
		t.Fatalf("expected start command in panel, got: %q", out)
	}
}
