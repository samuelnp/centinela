package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func verifyDepsRepo(t *testing.T, feature, contract string) {
	t.Helper()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	wf := workflow.New(feature)
	wf.PlanContract = contract
	if err := workflow.Save(wf); err != nil {
		t.Fatal(err)
	}
}

// D3: the shared constructor must resolve roles contract-awarely. Omitting
// Deps.Roles is silent — verify.Verify falls back to the contract-blind policy
// layer and reports "no claims to verify" for a legacy workflow.
func TestVerifyDepsFor_PinnedPlanResolvesPlanner(t *testing.T) {
	verifyDepsRepo(t, "pinned-workflow", workflow.PlanContractUnified)
	deps := verifyDepsFor("pinned-workflow", "plan")
	if len(deps.Roles) != 1 || string(deps.Roles[0]) != "planner" {
		t.Fatalf("pinned plan roles = %v, want [planner]", deps.Roles)
	}
}

func TestVerifyDepsFor_UnpinnedPlanResolvesLegacyPair(t *testing.T) {
	verifyDepsRepo(t, "unpinned-workflow", "")
	deps := verifyDepsFor("unpinned-workflow", "plan")
	if len(deps.Roles) != 2 {
		t.Fatalf("unpinned plan roles = %v, want the legacy pair", deps.Roles)
	}
	if string(deps.Roles[0]) != "big-thinker" || string(deps.Roles[1]) != "feature-specialist" {
		t.Fatalf("unpinned plan roles = %v, want [big-thinker feature-specialist]", deps.Roles)
	}
}

// Roles must never be nil: a nil slice is exactly the fallback that produced the
// false green, so an empty result would silently reintroduce it.
func TestVerifyDepsFor_RolesNeverNilForAGatedStep(t *testing.T) {
	verifyDepsRepo(t, "pinned-workflow", workflow.PlanContractUnified)
	for _, step := range []string{"plan", "code", "tests", "validate"} {
		if deps := verifyDepsFor("pinned-workflow", step); deps.Roles == nil {
			t.Errorf("step %q resolved nil Roles — verify would fall back to the policy layer", step)
		}
	}
}

func TestVerifyDepsFor_CarriesRootAndRunner(t *testing.T) {
	verifyDepsRepo(t, "pinned-workflow", workflow.PlanContractUnified)
	deps := verifyDepsFor("pinned-workflow", "plan")
	if deps.Root == "" {
		t.Error("Root must be set so verification runs against the worktree")
	}
	if deps.Runner == nil {
		t.Error("Runner must be set")
	}
}
