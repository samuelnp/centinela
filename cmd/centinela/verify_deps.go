package main

import (
	"github.com/samuelnp/centinela/internal/verify"
	"github.com/samuelnp/centinela/internal/workflow"
)

// verifyDepsFor builds the claim-verification dependencies for feature/step
// with the roles resolved CONTRACT-AWARELY.
//
// This exists as one constructor because omitting Roles is silent: verify.Verify
// falls back to orchestration.RequiredRoles(step), which cannot see a workflow's
// pinned PlanContract/ValidateContract, so a legacy workflow's evidence is looked
// up under the wrong role and the surface reports "no claims to verify" while
// `centinela verify` checks the same tree in full. That false green shipped on
// `centinela verdict` and the MCP verdict tool; routing every surface through
// here means a new caller gets the contract-aware behavior by default instead of
// having to remember a field.
//
// D3: the directive, the complete gate, `verify`, `complete_verify`, `verdict`
// and MCP must all name the SAME required set for the same workflow.
func verifyDepsFor(feature, step string) verify.Deps {
	return verify.Deps{
		Root:   verifyRoot(),
		Runner: verify.NewExecRunner(),
		Roles:  workflow.RequiredEvidenceRoles(feature, step),
	}
}
