package ui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// panelDietBudget caps the combined byte size of the three hook panels that
// dominate a plan-step UserPromptSubmit turn. Measured baseline 2026-07-30:
// 1195 bytes (2004 before this feature removed the lipgloss.JoinVertical
// right-padding). ~17% headroom absorbs a clarified sentence in a nudge panel;
// a reintroduced padding pass roughly doubles the count, so it still fails.
const panelDietBudget = 1400

// panelDietBaseline is the measurement the budget was derived from, quoted in
// the failure message so the reader sees drift, not just a bare number.
const panelDietBaseline = 1195

// panelDietPadEnv appends synthetic bytes to the measured string. It exists so
// the acceptance suite can prove this guard actually goes red past budget — a
// guard nobody has watched fail is not a guard. Test-only; no production code
// reads it, and an unset/zero/garbage value is a no-op.
const panelDietPadEnv = "CENTINELA_PANEL_DIET_PAD"

// panelDietRender renders the fixed measurement fixture from the plan: one
// synthetic in-memory workflow, canonical step order (StepOrder left nil so
// OrderedSteps falls back to DefaultStepOrder). Deliberately NOT keyed to live
// repo state — no filesystem read, no active-workflow count, no os.Chdir — so
// the number is identical on every machine and every worktree.
func panelDietRender() string {
	wf := &workflow.Workflow{Feature: "sample-feature", CurrentStep: "plan"}
	return RenderContextCapped([]*workflow.Workflow{wf}, 0) + "\n" +
		RenderReviewReady("sample-feature", "plan", "code") + "\n" +
		RenderFeatureBriefNeeded("sample-feature")
}

func panelDietMeasured() string {
	out := panelDietRender()
	if n, err := strconv.Atoi(os.Getenv(panelDietPadEnv)); err == nil && n > 0 {
		out += strings.Repeat("x", n)
	}
	return out
}

func TestPanelBudgetNoTrailingWhitespace(t *testing.T) {
	for i, line := range strings.Split(panelDietRender(), "\n") {
		if line != strings.TrimRight(line, " \t") {
			t.Fatalf("hook panel line %d ends in whitespace: %q\n"+
				"A hook render function lost its trimTrailingWS wrap — check the six "+
				"returns in internal/ui/render.go, render_review.go, render_brief.go.", i+1, line)
		}
	}
}

func TestPanelBudgetWithinByteBudget(t *testing.T) {
	out := panelDietMeasured()
	if n := len(out); n > panelDietBudget {
		t.Fatalf("hook panel budget exceeded: measured %d bytes, budget %d "+
			"(baseline %d, measured 2026-07-30).\n"+
			"If the copy legitimately grew: re-measure and raise panelDietBudget in "+
			"internal/ui/panel_budget_test.go with a dated note.\n"+
			"If it roughly doubled: lipgloss.JoinVertical padding is back — confirm the six "+
			"hook render functions still wrap their return in trimTrailingWS.",
			n, panelDietBudget, panelDietBaseline)
	}
}

// TestPanelBudgetIndependentOfRepoState is the machine-independence proof: the
// same measurement taken from a directory holding seven unrelated workflow
// state files must be byte-identical to the one taken from the package dir. A
// guard that read live state would differ here and flake as any repo grows.
func TestPanelBudgetIndependentOfRepoState(t *testing.T) {
	before := panelDietRender()

	dir := filepath.Join(t.TempDir(), ".workflow")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 7; i++ {
		name := "noise-feature-" + strconv.Itoa(i)
		body := `{"feature":"` + name + `","currentStep":"code","steps":{}}`
		if err := os.WriteFile(filepath.Join(dir, name+".json"), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(filepath.Dir(dir))

	if after := panelDietRender(); after != before {
		t.Fatalf("size-guard fixture changed with repo state: %d bytes here vs %d "+
			"at the package dir — the guard must measure the fixed in-memory fixture, "+
			"never the live active-workflow count.", len(after), len(before))
	}
}
