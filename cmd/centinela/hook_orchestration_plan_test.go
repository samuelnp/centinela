package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// setupPlanDirective writes a workflow with the given PlanContract and returns
// the hook directive it produces.
func setupPlanDirective(t *testing.T, contract string) string {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) })       //nolint:errcheck
	os.Chdir(d)                             //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0755) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = "plan"
	wf.PlanContract = contract
	workflow.Save(wf) //nolint:errcheck
	return captureStdout(t, func() {
		withStdin(t, "{}", func() { runHookOrchestration(nil, nil) }) //nolint:errcheck
	})
}

// D3: the directive names exactly one role and only the planner evidence pair.
func TestHookDirective_PinnedPlanNamesPlannerOnly(t *testing.T) {
	out := setupPlanDirective(t, workflow.PlanContractUnified)
	if !strings.Contains(out, "planner") {
		t.Fatalf("directive must name planner: %s", out)
	}
	for _, retired := range []string{"big-thinker", "feature-specialist"} {
		if strings.Contains(out, retired) {
			t.Fatalf("directive must not mention %q: %s", retired, out)
		}
	}
	if !strings.Contains(out, ".workflow/f-planner.md, .workflow/f-planner.json") {
		t.Fatalf("directive must list only the planner pair: %s", out)
	}
	if !strings.Contains(out, "reasoning") && !strings.Contains(out, "opus") {
		t.Fatalf("directive must carry the reasoning-tier model reference: %s", out)
	}
}

// D3: an unpinned workflow's directive names the legacy pair the gate enforces.
func TestHookDirective_UnpinnedPlanNamesLegacyPair(t *testing.T) {
	out := setupPlanDirective(t, "")
	for _, want := range []string{"big-thinker", "feature-specialist"} {
		if !strings.Contains(out, want) {
			t.Fatalf("directive must name %q for a legacy workflow: %s", want, out)
		}
	}
	if strings.Contains(out, "f-planner.md") {
		t.Fatalf("a legacy workflow must not be told to write planner evidence: %s", out)
	}
}

// D8 end-to-end: a legacy model-override key resolves for the planner role all
// the way from TOML to the directive's model reference.
func TestHookDirective_LegacyModelKeyAliasesToPlanner(t *testing.T) {
	out := setupRoutingRepo(t, "plan", "[orchestration.models]\nbig-thinker = \"fast\"\n")
	if !strings.Contains(out, "planner (model: haiku (claude)") {
		t.Fatalf("legacy big-thinker key must alias onto planner: %s", out)
	}
}

// D8: the deprecation notice never appears on the per-prompt hook path.
func TestHookDirective_NeverRepeatsDeprecationNotice(t *testing.T) {
	out := setupRoutingRepo(t, "plan", "[orchestration.models]\nbig-thinker = \"fast\"\n")
	if strings.Contains(out, "retired plan role key") {
		t.Fatalf("hook output must not carry the migration notice: %s", out)
	}
}
