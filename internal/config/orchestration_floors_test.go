package config

import (
	"strings"
	"testing"
)

func floorsConfig(floors map[string]string) *Config {
	cfg := &Config{}
	cfg.Orchestration.Floors = floors
	return cfg
}

func TestOrchestrationFloors_NormalizesAndAliasesPlanRole(t *testing.T) {
	if OrchestrationFloors(nil) != nil || OrchestrationFloors(&Config{}) != nil {
		t.Fatal("an absent floors table must return nil")
	}
	got := OrchestrationFloors(floorsConfig(map[string]string{" GATEKEEPER ": " Reasoning "}))
	if got["gatekeeper"] != "reasoning" {
		t.Fatalf("keys and values must be trimmed + lowercased, got %#v", got)
	}
	aliased := OrchestrationFloors(floorsConfig(map[string]string{"big-thinker": "reasoning"}))
	if aliased["planner"] != "reasoning" {
		t.Fatalf("a retired plan key must alias onto planner, got %#v", aliased)
	}
	explicit := OrchestrationFloors(floorsConfig(map[string]string{
		"big-thinker": "reasoning", "planner": "fast",
	}))
	if explicit["planner"] != "fast" {
		t.Fatalf("an explicit planner floor must win over the alias, got %#v", explicit)
	}
}

func TestValidateOrchestrationFloors(t *testing.T) {
	if err := validateOrchestrationFloors(floorsConfig(map[string]string{"gatekeeper": "reasoning"})); err != nil {
		t.Fatalf("a valid floor must be accepted: %v", err)
	}
	err := validateOrchestrationFloors(floorsConfig(map[string]string{"wizard": "fast"}))
	if err == nil || !strings.Contains(err.Error(), "wizard") {
		t.Fatalf("an unknown role key must be refused naming it, got %v", err)
	}
	err = validateOrchestrationFloors(floorsConfig(map[string]string{"qa-senior": "turbo"}))
	if err == nil || !strings.Contains(err.Error(), "reasoning, balanced, fast") {
		t.Fatalf("an unknown tier must be refused listing the tiers, got %v", err)
	}
}

// One vocabulary, two validators: a key [orchestration.models] rejects must not
// be accepted — and silently applied — by [orchestration.floors].
func TestValidateOrchestrationFloors_KeyCasingMatchesModels(t *testing.T) {
	for _, key := range []string{"Gatekeeper", "GATEKEEPER", " gatekeeper "} {
		cfg := floorsConfig(map[string]string{key: "reasoning"})
		cfg.Orchestration.Models = map[string]RoleModelValue{key: {Tier: "reasoning"}}
		floorsErr := validateOrchestrationFloors(cfg)
		modelsErr := validateOrchestrationModels(cfg)
		if (floorsErr == nil) != (modelsErr == nil) {
			t.Fatalf("key %q disagrees: floors=%v models=%v", key, floorsErr, modelsErr)
		}
	}
}
