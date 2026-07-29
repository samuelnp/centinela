package config

import (
	"strings"
	"testing"
)

func modelsCfg(entries map[string]RoleModelValue) *Config {
	c := &Config{}
	c.Orchestration.Models = entries
	return c
}

func TestAliasPlanRole_BigThinkerTierPublishedAsPlanner(t *testing.T) {
	cfg := modelsCfg(map[string]RoleModelValue{"big-thinker": {Tier: "fast"}})
	tiers := OrchestrationModelTiers(cfg)
	if tiers["planner"] != "fast" {
		t.Fatalf("planner tier = %q, want the aliased big-thinker value", tiers["planner"])
	}
	if tiers["big-thinker"] != "fast" {
		t.Fatal("the legacy key must stay in the returned map")
	}
}

func TestAliasPlanRole_ExplicitPlannerWins(t *testing.T) {
	cfg := modelsCfg(map[string]RoleModelValue{
		"planner":     {Tier: "reasoning"},
		"big-thinker": {Tier: "fast"},
	})
	if got := OrchestrationModelTiers(cfg)["planner"]; got != "reasoning" {
		t.Fatalf("explicit planner must win, got %q", got)
	}
}

// big-thinker set the tier planner inherits, so it beats feature-specialist.
func TestAliasPlanRole_BigThinkerBeatsFeatureSpecialist(t *testing.T) {
	cfg := modelsCfg(map[string]RoleModelValue{
		"big-thinker":        {Tier: "reasoning"},
		"feature-specialist": {Tier: "balanced"},
	})
	if got := OrchestrationModelTiers(cfg)["planner"]; got != "reasoning" {
		t.Fatalf("big-thinker must win precedence, got %q", got)
	}
}

func TestAliasPlanRole_FeatureSpecialistAloneAliases(t *testing.T) {
	cfg := modelsCfg(map[string]RoleModelValue{"feature-specialist": {Tier: "balanced"}})
	if got := OrchestrationModelTiers(cfg)["planner"]; got != "balanced" {
		t.Fatalf("feature-specialist alone must alias, got %q", got)
	}
}

func TestAliasPlanRole_OverrideTableFormAliasesToo(t *testing.T) {
	cfg := modelsCfg(map[string]RoleModelValue{
		"big-thinker": {Overrides: map[string]string{"claude": "some-model"}},
	})
	got := OrchestrationModelOverrides(cfg)["planner"]
	if got["claude"] != "some-model" {
		t.Fatalf("override table did not alias to planner: %#v", got)
	}
}

// An explicit planner tier must not be shadowed by an aliased legacy override.
func TestAliasPlanRole_ExplicitPlannerSuppressesOverrideAlias(t *testing.T) {
	cfg := modelsCfg(map[string]RoleModelValue{
		"planner":     {Tier: "fast"},
		"big-thinker": {Overrides: map[string]string{"claude": "some-model"}},
	})
	if _, ok := OrchestrationModelOverrides(cfg)["planner"]; ok {
		t.Fatal("an explicit planner entry must suppress the override alias")
	}
}

func TestAliasPlanRole_NoLegacyKeysIsNoop(t *testing.T) {
	cfg := modelsCfg(map[string]RoleModelValue{"senior-engineer": {Tier: "fast"}})
	if _, ok := OrchestrationModelTiers(cfg)["planner"]; ok {
		t.Fatal("planner must not be invented when no plan key is configured")
	}
	if OrchestrationModelTiers(nil) != nil {
		t.Fatal("nil config must stay nil")
	}
}

func TestLegacyPlanModelRoleKeys_SortedAndNilSafe(t *testing.T) {
	if LegacyPlanModelRoleKeys(nil) != nil {
		t.Fatal("nil config must yield no keys")
	}
	if got := LegacyPlanModelRoleKeys(modelsCfg(nil)); got != nil {
		t.Fatalf("empty models must yield no keys, got %v", got)
	}
	cfg := modelsCfg(map[string]RoleModelValue{
		"feature-specialist": {Tier: "balanced"},
		"big-thinker":        {Tier: "reasoning"},
		"senior-engineer":    {Tier: "fast"},
	})
	got := LegacyPlanModelRoleKeys(cfg)
	if strings.Join(got, ",") != "big-thinker,feature-specialist" {
		t.Fatalf("keys = %v, want the two legacy keys in precedence order", got)
	}
}
