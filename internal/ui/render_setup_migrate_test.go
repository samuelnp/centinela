package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/setup"
)

func TestRenderSetupMigrationOutputs(t *testing.T) {
	plan := setup.SyncPlan{Items: []setup.SyncItem{{
		Path: "AGENTS.md", Action: setup.SyncUpdate, Reason: "managed version bump",
	}}}
	out := RenderSetupMigrationPlan(plan, false)
	if !strings.Contains(out, "SETUP PREVIEW") || !strings.Contains(out, "AGENTS.md") {
		t.Fatal("expected setup preview details")
	}
	needed := RenderMigrationNeeded(2, 3)
	if !strings.Contains(needed, "MIGRATION REQUIRED") || !strings.Contains(needed, "centinela migrate --apply") {
		t.Fatal("expected unified migration-needed guidance")
	}
}
