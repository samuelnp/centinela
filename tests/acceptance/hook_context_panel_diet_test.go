// Acceptance: specs/hook-context-panel-diet.feature
package acceptance_test

import (
	"os/exec"
	"strings"
	"testing"
)

// Scenario: Hook context output carries no trailing whitespace
func TestPDHookContextHasNoTrailingWhitespace(t *testing.T) {
	bin := pdBuildBin(t)
	dir := pdRepo(t)
	pdWorkflow(t, dir, "sample-feature", "plan", pdCanonical)
	pdPlanArtifacts(t, dir, "sample-feature")

	out := pdHookContext(t, bin, dir)
	mustContain(t, out, "ACTIVE WORKFLOWS")
	pdRequireUnpadded(t, "hook context", out)
}

// Scenario: A feature-brief-required nudge panel carries no trailing whitespace
func TestPDFeatureBriefPanelHasNoTrailingWhitespace(t *testing.T) {
	bin := pdBuildBin(t)
	dir := pdRepo(t)
	// No docs/features/sample-feature.md, so the brief nudge fires.
	pdWorkflow(t, dir, "sample-feature", "plan", pdCanonical)

	out := pdHookContext(t, bin, dir)
	mustContain(t, out, "FEATURE BRIEF REQUIRED")
	pdRequireUnpadded(t, "feature-brief-required panel", out)
}

// Scenario: A review-required nudge panel carries no trailing whitespace
func TestPDReviewReadyPanelHasNoTrailingWhitespace(t *testing.T) {
	out := pdReviewReadyOutput(t)
	mustContain(t, out, "REVIEW REQUIRED")
	pdRequireUnpadded(t, "review-required panel", out)
}

// pdReviewReadyOutput renders hook context for a plan step whose artifacts all
// validate, which is the condition that arms the REVIEW REQUIRED panel.
func pdReviewReadyOutput(t *testing.T) string {
	t.Helper()
	bin := pdBuildBin(t)
	dir := pdRepo(t)
	pdWorkflow(t, dir, "sample-feature", "plan", pdCanonical)
	pdPlanArtifacts(t, dir, "sample-feature")
	return pdHookContext(t, bin, dir)
}

// Scenario: The active feature, its step, and its progress still appear
func TestPDGovernanceHeaderSurvivesTheDiet(t *testing.T) {
	bin := pdBuildBin(t)
	dir := pdRepo(t)
	pdWorkflow(t, dir, "sample-feature", "plan", pdCanonical)
	pdPlanArtifacts(t, dir, "sample-feature")

	out := pdHookContext(t, bin, dir)
	for _, want := range []string{"sample-feature", "plan", "0/5"} {
		mustContain(t, out, want)
	}
}

// Scenario: The full step ladder still distinguishes archetypes with different gates
func TestPDStepLadderStillDistinguishesArchetypes(t *testing.T) {
	bin := pdBuildBin(t)
	dir := pdRepo(t)
	pdWorkflow(t, dir, "sample-feature", "plan", pdCanonical)
	pdWorkflow(t, dir, "quick-check", "code", []string{"plan", "code"})

	out := pdHookContext(t, bin, dir)
	ladders := pdLadders(t, out, "sample-feature", "quick-check")
	if strings.Contains(ladders["quick-check"], "validate") {
		t.Errorf("spike ladder must not advertise a validate step: %q", ladders["quick-check"])
	}
	if !strings.Contains(ladders["sample-feature"], "validate") {
		t.Errorf("canonical ladder lost its validate step: %q", ladders["sample-feature"])
	}
}

// pdLadders maps each feature name to the step-ladder line rendered under it.
func pdLadders(t *testing.T, out string, features ...string) map[string]string {
	t.Helper()
	lines := strings.Split(out, "\n")
	got := map[string]string{}
	for i, line := range lines {
		for _, f := range features {
			if strings.Contains(line, f) && i+1 < len(lines) && strings.Contains(lines[i+1], "·") {
				got[f] = lines[i+1]
			}
		}
	}
	if len(got) != len(features) {
		t.Fatalf("expected one ladder per workflow, got %d:\n%s", len(got), out)
	}
	return got
}

// Scenario: A blocking review-required instruction still appears when due
func TestPDReviewInstructionStillNamesItsFeature(t *testing.T) {
	out := pdReviewReadyOutput(t)
	mustContain(t, out, "STOP. Do not advance.")
	// The nudge loop is flat across all active workflows, so the panel's own
	// self-identification is the only way to know which feature it is about.
	mustContain(t, out, "sample-feature · plan artifacts complete")
}

// Scenario: centinela status output is unchanged, trailing whitespace included
func TestPDStatusOutputKeepsItsPadding(t *testing.T) {
	bin := pdBuildBin(t)
	dir := pdRepo(t)
	pdWorkflow(t, dir, "sample-feature", "plan", pdCanonical)

	out, code := runCent(t, bin, dir, "status", "sample-feature")
	if code != 0 {
		t.Fatalf("status exited %d: %s", code, out)
	}
	mustContain(t, out, "sample-feature")
	if len(pdPaddedLines(out)) == 0 {
		t.Fatalf("centinela status lost its padding — this feature changed what the HOOK "+
			"emits, not how the CLI renders to a TTY:\n%s", out)
	}
}

// Scenario: A blocked-write refusal panel is unaffected by this feature
func TestPDBlockedWritePanelKeepsItsPadding(t *testing.T) {
	bin := pdBuildBin(t)
	dir := pdRepo(t)
	pdWorkflow(t, dir, "sample-feature", "plan", pdCanonical)

	c := exec.Command(bin, "hook", "prewrite")
	c.Dir = dir
	c.Stdin = strings.NewReader(`{"tool_input":{"file_path":"` + dir + `/internal/thing.go"}}`)
	raw, _ := c.CombinedOutput()
	out := string(raw)
	mustContain(t, out, "BLOCKED WRITE")
	if len(pdPaddedLines(out)) == 0 {
		t.Fatalf("blocked-write refusal lost its padding — RenderBlocked is out of this "+
			"feature's scope and must be byte-identical:\n%s", out)
	}
}
