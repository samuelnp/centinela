package unit_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// AC1: tier-map override for the active runner is used.
func TestRouting_ResolveTierMapForRunner(t *testing.T) {
	mm := orchestration.ModelMap{"reasoning": {"opencode": "moonshotai/kimi-k2"}}
	got, ok := orchestration.ResolveModel(orchestration.RoleBigThinker, nil, mm, orchestration.RunnerOpenCode)
	if !ok || got != "moonshotai/kimi-k2" {
		t.Errorf("AC1: got (%q, %v)", got, ok)
	}
}

// AC2: role override beats the tier (and any model_map entry).
func TestRouting_ResolveRoleOverrideWins(t *testing.T) {
	models := orchestration.RoleModels{"senior-engineer": {Overrides: map[string]string{"opencode": "deepseek/deepseek-coder"}}}
	mm := orchestration.ModelMap{"reasoning": {"opencode": "moonshotai/kimi-k2"}}
	got, ok := orchestration.ResolveModel(orchestration.RoleSeniorEngineer, models, mm, orchestration.RunnerOpenCode)
	if !ok || got != "deepseek/deepseek-coder" {
		t.Errorf("AC2: override should win, got (%q, %v)", got, ok)
	}
}

// AC3: tier override but no model_map entry for the runner → built-in default.
func TestRouting_ResolveNoMapEntryUsesDefault(t *testing.T) {
	models := orchestration.RoleModels{"feature-specialist": {Tier: "balanced"}}
	mm := orchestration.ModelMap{"balanced": {"opencode": "deepseek/deepseek-chat"}}
	got, ok := orchestration.ResolveModel(orchestration.RoleFeatureSpecial, models, mm, orchestration.RunnerClaude)
	if !ok || got != "claude-sonnet-4-6" {
		t.Errorf("AC3: got (%q, %v)", got, ok)
	}
}

// AC7: codex (empty column) returns the tier name + ok=false, no leak.
func TestRouting_ResolveCodexFallback(t *testing.T) {
	mm := orchestration.ModelMap{"reasoning": {"opencode": "moonshotai/kimi-k2"}}
	got, ok := orchestration.ResolveModel(orchestration.RoleBigThinker, nil, mm, orchestration.RunnerCodex)
	if ok || got != "reasoning" {
		t.Errorf("AC7: got (%q, %v)", got, ok)
	}
}
