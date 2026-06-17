package orchestration

import (
	"strings"
	"testing"
)

func TestResolveModel_DefaultAndOverride(t *testing.T) {
	// Default: big-thinker → reasoning → claude-opus-4-7 (claude)
	got, ok := ResolveModel(RoleBigThinker, nil, nil, RunnerClaude)
	if !ok || got != "claude-opus-4-7" {
		t.Errorf("expected (claude-opus-4-7, true), got (%q, %v)", got, ok)
	}
	// Config override: big-thinker = fast → claude-haiku
	models := RoleModels{"big-thinker": {Tier: "fast"}}
	got, ok = ResolveModel(RoleBigThinker, models, nil, RunnerClaude)
	if !ok || got != "claude-haiku-4-5-20251001" {
		t.Errorf("override: expected (claude-haiku-4-5-20251001, true), got (%q, %v)", got, ok)
	}
}

func TestResolveModel_OpenCodeRunner(t *testing.T) {
	cases := []struct{ tier, want string }{
		{"reasoning", "anthropic/claude-opus-4-7"},
		{"balanced", "anthropic/claude-sonnet-4-6"},
		{"fast", "anthropic/claude-haiku-4-5"},
	}
	for _, tc := range cases {
		models := RoleModels{"big-thinker": {Tier: tc.tier}}
		got, ok := ResolveModel(RoleBigThinker, models, nil, RunnerOpenCode)
		if !ok || got != tc.want {
			t.Errorf("opencode %q: expected (%q, true), got (%q, %v)", tc.tier, tc.want, got, ok)
		}
	}
}

func TestResolveModel_UnknownRunnerFallback(t *testing.T) {
	got, ok := ResolveModel(RoleBigThinker, nil, nil, RunnerUnknown)
	if ok {
		t.Errorf("expected ok=false for unknown runner, got true; model=%q", got)
	}
	if got != "reasoning" {
		t.Errorf("expected tier-name fallback 'reasoning', got %q", got)
	}
}

func TestResolveModel_NilMapNoPanic(t *testing.T) {
	got, ok := ResolveModel(RoleDocsSpecialist, nil, nil, RunnerClaude)
	if !ok || got != "claude-haiku-4-5-20251001" {
		t.Errorf("nil map: expected (claude-haiku-4-5-20251001, true), got (%q, %v)", got, ok)
	}
}

func TestModelReference_AllTiers(t *testing.T) {
	ref := ModelReference([]Tier{TierReasoning, TierBalanced, TierFast})
	for _, want := range []string{
		"claude-opus-4-7", "anthropic/claude-opus-4-7",
		"claude-sonnet-4-6", "anthropic/claude-sonnet-4-6",
		"claude-haiku-4-5-20251001", "anthropic/claude-haiku-4-5",
	} {
		if !strings.Contains(ref, want) {
			t.Errorf("ModelReference missing %q; got: %s", want, ref)
		}
	}
}

func TestModelReference_Dedup(t *testing.T) {
	ref := ModelReference([]Tier{TierReasoning, TierReasoning})
	if strings.Count(ref, "reasoning:") != 1 {
		t.Errorf("expected one reasoning entry, got: %s", ref)
	}
}

func TestModelReference_EmptySlice(t *testing.T) {
	ref := ModelReference(nil)
	if ref != "" {
		t.Errorf("expected empty string for nil tiers, got %q", ref)
	}
}

func TestContainsTier(t *testing.T) {
	tiers := []Tier{TierReasoning, TierFast}
	if !containsTier(tiers, TierReasoning) {
		t.Error("expected containsTier to find TierReasoning")
	}
	if containsTier(tiers, TierBalanced) {
		t.Error("expected containsTier to NOT find TierBalanced")
	}
}
