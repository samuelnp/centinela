// Acceptance: specs/token-diet.feature
package acceptance_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

var tdDatedSnapshot = regexp.MustCompile(`-[0-9]{8}\b`)

// tdDirective runs `hook orchestration` in dir and returns its output.
func tdDirective(t *testing.T, bin, dir string) string {
	t.Helper()
	out, code := runCent(t, bin, dir, "hook", "orchestration")
	if code != 0 {
		t.Fatalf("hook orchestration failed: %s", out)
	}
	return out
}

// Scenario: No built-in model default is a dated snapshot
func TestTD_NoBuiltinModelDefaultIsDatedSnapshot(t *testing.T) {
	for _, id := range orchestration.TierModelIDs() {
		if tdDatedSnapshot.MatchString(id) {
			t.Fatalf("built-in id %q looks like a dated snapshot", id)
		}
	}
}

// Scenario: The claude runner's defaults are bare family aliases
func TestTD_ClaudeRunnerDefaultsAreBareFamilyAliases(t *testing.T) {
	cases := []struct {
		role  orchestration.Role
		alias string
	}{
		{orchestration.RolePlanner, "opus"},          // reasoning
		{orchestration.RoleFeatureSpecial, "sonnet"}, // balanced
		{orchestration.RoleDocsSpecialist, "haiku"},  // fast
	}
	for _, tc := range cases {
		id, ok := orchestration.ResolveModel(tc.role, nil, nil, orchestration.RunnerClaude)
		if !ok || id != tc.alias {
			t.Fatalf("role %s: got (%q, %v), want (%q, true)", tc.role, id, ok, tc.alias)
		}
		if strings.ContainsAny(id, "0123456789") {
			t.Fatalf("alias %q must contain no digits", id)
		}
	}
}

// Scenario: A tier remap in config overrides the built-in alias
func TestTD_TierRemapInConfigOverridesBuiltinAlias(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdWriteWorkflow(t, dir, "token-diet", "plan", "planner-v1", true)
	mustWrite(t, dir+"/centinela.toml", "[orchestration.model_map.reasoning]\nclaude = \"custom-opus-x1\"\n")

	out := tdDirective(t, bin, dir)
	mustContain(t, out, "model: custom-opus-x1 (claude)")
	// Precedence otherwise unchanged: opencode still built-in, codex still rule-4.
	mustContain(t, out, "model: anthropic/claude-opus-4-7 (opencode)")
	mustContain(t, out, "model: reasoning (codex)")
}

// Scenario: A runner with no built-in entry still falls back to the tier name
func TestTD_RunnerWithNoBuiltinEntryFallsBackToTierName(t *testing.T) {
	id, ok := orchestration.ResolveModel(orchestration.RolePlanner, nil, nil, orchestration.RunnerCodex)
	if ok {
		t.Fatalf("codex has no built-in reasoning entry; ok must be false, got id=%q", id)
	}
	if id != "reasoning" {
		t.Fatalf("fallback value must be the tier name, got %q", id)
	}
	if id == "opus" || id == "anthropic/claude-opus-4-7" {
		t.Fatal("fallback must never leak another runner's concrete model id")
	}
}

// Scenario: The orchestration directive prints the alias, not a dated pin
func TestTD_DirectivePrintsAliasNotDatedPin(t *testing.T) {
	bin := tdBuildBin(t)
	dir := tdRepo(t)
	tdWriteWorkflow(t, dir, "token-diet", "plan", "planner-v1", true)

	out := tdDirective(t, bin, dir)
	mustContain(t, out, "planner")
	mustContain(t, out, "model: opus (claude)")
	mustContain(t, out, "model: reasoning (codex)")
	if tdDatedSnapshot.MatchString(out) {
		t.Fatalf("directive must contain no dated snapshot id: %s", out)
	}
}
