package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// RenderStatus surfaces the pinned archetype so the active track is visible. A
// spike additionally carries the no-ship-gate annotation; an unpinned workflow
// shows canonical.
func TestRenderStatus_ArchetypeLine(t *testing.T) {
	spike := &workflow.Workflow{Feature: "f", CurrentStep: "code", Archetype: workflow.ArchetypeSpike}
	out := RenderStatus(spike)
	if !strings.Contains(out, "Archetype") || !strings.Contains(out, "spike") {
		t.Fatalf("status must show the spike archetype, got:\n%s", out)
	}
	if !strings.Contains(out, "no ship gate") {
		t.Fatalf("spike must carry the no-ship-gate annotation, got:\n%s", out)
	}

	bare := &workflow.Workflow{Feature: "f", CurrentStep: "plan"}
	if !strings.Contains(RenderStatus(bare), workflow.ArchetypeCanonical) {
		t.Fatalf("unpinned workflow must show canonical, got:\n%s", RenderStatus(bare))
	}
}
