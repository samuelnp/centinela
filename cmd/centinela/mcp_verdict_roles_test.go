package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// D3, unpinned row: MCP must verify the legacy pair, not the policy default.
func TestMCPVerdict_VerifiesLegacyPlanRoles(t *testing.T) {
	mcpRepo(t, "legacy-feature", "")
	for _, r := range []string{"big-thinker", "feature-specialist"} {
		mcpEvidence(t, "legacy-feature", "plan", r)
	}
	got := mcpVerifiedRoles(t, "legacy-feature")
	if len(got) != 2 || got[0] != "big-thinker" || got[1] != "feature-specialist" {
		t.Fatalf("MCP verified %v, want exactly [big-thinker feature-specialist]", got)
	}
}

// D3, pinned row: asserted POSITIVELY. Planner evidence exists AND both retired
// roles' evidence exists, so the exact set is discriminating in both directions.
func TestMCPVerdict_VerifiesPinnedPlannerRole(t *testing.T) {
	mcpRepo(t, "pinned-feature", workflow.PlanContractUnified)
	for _, r := range []string{"planner", "big-thinker", "feature-specialist"} {
		mcpEvidence(t, "pinned-feature", "plan", r)
	}
	got := mcpVerifiedRoles(t, "pinned-feature")
	if len(got) != 1 || got[0] != "planner" {
		t.Fatalf("MCP verified %v, want exactly [planner]", got)
	}
}
