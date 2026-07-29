package main

import (
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// TestMCPVerdict_UsesTheContractAwareResolverNotThePolicyLayer is the
// drop-Roles kill for the MCP surface.
//
// At the PINNED plan row the two resolvers agree by construction —
// orchestration.RequiredRoles("plan") and workflow.RequiredEvidenceRoles both
// yield [planner] — so no pinned/plan assertion can discriminate a missing
// Deps.Roles. The code step of a USER-FACING feature is where they diverge:
// RequiredRolesForFeature appends ux-ui-specialist, the bare policy fallback
// does not. So if mcp.go ever stops passing Roles, ux-ui-specialist silently
// goes unverified and this test fails.
func TestMCPVerdict_UsesTheContractAwareResolverNotThePolicyLayer(t *testing.T) {
	const feature = "surface-feature"
	mcpRepo(t, feature, workflow.PlanContractUnified)
	// Declare the feature user-facing: this is what makes the feature-aware
	// resolver add ux-ui-specialist at the code step.
	mcpWrite(t, filepath.Join("docs/features", feature+".md"),
		"# brief\n\nsurface: user-facing\n")
	wf, err := workflow.Load(feature)
	if err != nil {
		t.Fatal(err)
	}
	wf.CurrentStep = "code"
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
	mcpEvidence(t, feature, "code", "senior-engineer")
	mcpEvidence(t, feature, "code", "ux-ui-specialist")

	got := mcpVerifiedRoles(t, feature)
	seen := map[string]bool{}
	for _, r := range got {
		seen[r] = true
	}
	if !seen["senior-engineer"] {
		t.Fatalf("MCP verified %v, missing senior-engineer", got)
	}
	if !seen["ux-ui-specialist"] {
		t.Fatalf("MCP verified %v, missing ux-ui-specialist — the verdict surface "+
			"fell back to the contract-blind policy layer (Deps.Roles omitted)", got)
	}
}
