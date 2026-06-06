package orchestration

import "testing"

// Step 1 precedence: a role-level override for the runner beats everything.
func TestResolveModel_RoleOverrideWins(t *testing.T) {
	models := RoleModels{"senior-engineer": {Overrides: map[string]string{"opencode": "deepseek/deepseek-coder"}}}
	mm := ModelMap{"reasoning": {"opencode": "moonshotai/kimi-k2"}}
	got, ok := ResolveModel(RoleSeniorEngineer, models, mm, RunnerOpenCode)
	if !ok || got != "deepseek/deepseek-coder" {
		t.Errorf("override should win: got (%q, %v)", got, ok)
	}
}

// Step 1 ignores an empty override string and falls through to the tier path.
func TestResolveModel_EmptyOverrideFallsThrough(t *testing.T) {
	models := RoleModels{"big-thinker": {Overrides: map[string]string{"claude": ""}}}
	got, ok := ResolveModel(RoleBigThinker, models, nil, RunnerClaude)
	if !ok || got != "claude-opus-4-7" {
		t.Errorf("empty override should fall through: got (%q, %v)", got, ok)
	}
}

// Step 2 precedence: tier-map override for the runner beats the built-in default.
func TestResolveModel_TierMapOverride(t *testing.T) {
	mm := ModelMap{"reasoning": {"opencode": "moonshotai/kimi-k2"}}
	got, ok := ResolveModel(RoleBigThinker, nil, mm, RunnerOpenCode)
	if !ok || got != "moonshotai/kimi-k2" {
		t.Errorf("tier-map override expected kimi: got (%q, %v)", got, ok)
	}
}

// Step 3: a tier override with no model_map entry for the runner uses the default.
func TestResolveModel_TierOverrideNoMapEntry(t *testing.T) {
	models := RoleModels{"feature-specialist": {Tier: "balanced"}}
	mm := ModelMap{"balanced": {"opencode": "deepseek/deepseek-chat"}}
	got, ok := ResolveModel(RoleFeatureSpecial, models, mm, RunnerClaude)
	if !ok || got != "claude-sonnet-4-6" {
		t.Errorf("no claude map entry → built-in default: got (%q, %v)", got, ok)
	}
}

// Step 4: codex (empty column) returns the tier name + ok=false, never a leak.
func TestResolveModel_CodexRule4NoLeak(t *testing.T) {
	mm := ModelMap{"reasoning": {"opencode": "moonshotai/kimi-k2"}}
	got, ok := ResolveModel(RoleBigThinker, nil, mm, RunnerCodex)
	if ok || got != "reasoning" {
		t.Errorf("codex should fall to tier name: got (%q, %v)", got, ok)
	}
	if got == "moonshotai/kimi-k2" || got == "claude-opus-4-7" {
		t.Errorf("codex must not leak another runner's ID: %q", got)
	}
}

func TestRoleTier_OverrideAndDefault(t *testing.T) {
	if got := RoleTier(RoleBigThinker, nil); got != TierReasoning {
		t.Errorf("nil models → default reasoning, got %q", got)
	}
	models := RoleModels{"big-thinker": {Tier: "Fast"}}
	if got := RoleTier(RoleBigThinker, models); got != TierFast {
		t.Errorf("explicit Fast override should normalize to fast, got %q", got)
	}
	bad := RoleModels{"big-thinker": {Tier: "nonsense"}}
	if got := RoleTier(RoleBigThinker, bad); got != TierReasoning {
		t.Errorf("invalid tier override → default, got %q", got)
	}
}

func TestAllowedRunnerKeys_ThreeStable(t *testing.T) {
	keys := AllowedRunnerKeys()
	want := []string{"claude", "opencode", "codex"}
	if len(keys) != len(want) {
		t.Fatalf("expected %v, got %v", want, keys)
	}
	for i, k := range want {
		if keys[i] != k {
			t.Errorf("key %d = %q, want %q", i, keys[i], k)
		}
	}
}
