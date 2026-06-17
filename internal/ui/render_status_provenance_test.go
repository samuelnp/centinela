package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// RenderStatusWithConfig threads cfg so the Profile row shows full provenance:
// a frontier driver model annotates its derived class; zero-config shows default.
func TestRenderStatusWithConfig_Provenance(t *testing.T) {
	frontier := &workflow.Workflow{Feature: "f", CurrentStep: "code", DriverModel: "claude-opus-4-7"}
	out := RenderStatusWithConfig(frontier, &config.Config{})
	if !strings.Contains(out, config.ProfileOutcome) ||
		!strings.Contains(out, "driver: claude-opus-4-7 → frontier") {
		t.Fatalf("frontier driver provenance missing, got:\n%s", out)
	}

	zero := &workflow.Workflow{Feature: "f", CurrentStep: "plan"}
	zout := RenderStatusWithConfig(zero, &config.Config{})
	if !strings.Contains(zout, config.ProfileStrict) || !strings.Contains(zout, "(default)") {
		t.Fatalf("zero-config provenance missing, got:\n%s", zout)
	}
}

// The Worktree row renders only when a worktree path is pinned (covers the branch).
func TestRenderStatusWithConfig_WorktreeRow(t *testing.T) {
	wf := &workflow.Workflow{Feature: "f", CurrentStep: "code", WorktreePath: ".worktrees/f"}
	out := RenderStatusWithConfig(wf, nil)
	if !strings.Contains(out, "Worktree") || !strings.Contains(out, ".worktrees/f") {
		t.Fatalf("worktree row missing, got:\n%s", out)
	}
	if strings.Contains(RenderStatus(wf), "Worktree") != true {
		t.Fatal("worktree row should still appear without config")
	}
}
