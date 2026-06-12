package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// RenderStatus surfaces the effective (pinned) profile so it is visible which
// rails are active. A pinned profile shows verbatim; an empty one shows strict.
func TestRenderStatus_ProfileLine(t *testing.T) {
	pinned := &workflow.Workflow{Feature: "f", CurrentStep: "code", EnforcementProfile: config.ProfileOutcome}
	out := RenderStatus(pinned)
	if !strings.Contains(out, "Profile") || !strings.Contains(out, config.ProfileOutcome) {
		t.Fatalf("status must show pinned outcome profile, got:\n%s", out)
	}

	bare := &workflow.Workflow{Feature: "f", CurrentStep: "plan"}
	if !strings.Contains(RenderStatus(bare), config.ProfileStrict) {
		t.Fatalf("status must show strict for an unpinned workflow, got:\n%s", RenderStatus(bare))
	}
}
