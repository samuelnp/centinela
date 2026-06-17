package unit_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/orchestration"
)

func TestResolveModel_ConfigOverrideBeatsDefault(t *testing.T) {
	models := orchestration.RoleModels{"big-thinker": {Tier: "fast"}}
	got, ok := orchestration.ResolveModel(orchestration.RoleBigThinker, models, nil, orchestration.RunnerClaude)
	if !ok || got != "claude-haiku-4-5-20251001" {
		t.Errorf("expected claude-haiku-4-5-20251001, got (%q, %v)", got, ok)
	}
}

func TestResolveModel_ExactIDsPerRunner(t *testing.T) {
	cases := []struct {
		role   orchestration.Role
		runner orchestration.Runner
		want   string
	}{
		{orchestration.RoleBigThinker, orchestration.RunnerClaude, "claude-opus-4-7"},
		{orchestration.RoleBigThinker, orchestration.RunnerOpenCode, "anthropic/claude-opus-4-7"},
		{orchestration.RoleQASeniorEngineer, orchestration.RunnerClaude, "claude-sonnet-4-6"},
		{orchestration.RoleQASeniorEngineer, orchestration.RunnerOpenCode, "anthropic/claude-sonnet-4-6"},
		{orchestration.RoleDocsSpecialist, orchestration.RunnerClaude, "claude-haiku-4-5-20251001"},
		{orchestration.RoleDocsSpecialist, orchestration.RunnerOpenCode, "anthropic/claude-haiku-4-5"},
	}
	for _, tc := range cases {
		got, ok := orchestration.ResolveModel(tc.role, nil, nil, tc.runner)
		if !ok || got != tc.want {
			t.Errorf("ResolveModel(%q, nil, %q) = (%q, %v), want (%q, true)",
				tc.role, tc.runner, got, ok, tc.want)
		}
	}
}

func TestResolveModel_UnknownRunnerReturnsTierAndFalse(t *testing.T) {
	got, ok := orchestration.ResolveModel(orchestration.RoleBigThinker, nil, nil, orchestration.RunnerUnknown)
	if ok {
		t.Errorf("expected ok=false for unknown runner, got true; model=%q", got)
	}
	if got != "reasoning" {
		t.Errorf("expected tier name as fallback, got %q", got)
	}
}

func TestResolveModel_MissingMappingNoPanic(t *testing.T) {
	// RunnerUnknown has no entry in the per-runner table → returns tier name + ok=false
	got, ok := orchestration.ResolveModel(orchestration.RoleDocsSpecialist, nil, nil, orchestration.RunnerUnknown)
	if ok {
		t.Errorf("expected ok=false for unknown runner, got true; model=%q", got)
	}
	if got == "" {
		t.Error("expected non-empty tier-name fallback, got empty string")
	}
}

func TestResolveModel_NilModelsMapNoPanic(t *testing.T) {
	// Passing nil for models must not panic — falls back to default tier.
	got, ok := orchestration.ResolveModel(orchestration.RoleBigThinker, nil, nil, orchestration.RunnerClaude)
	if !ok || got != "claude-opus-4-7" {
		t.Errorf("nil models map: expected (claude-opus-4-7, true), got (%q, %v)", got, ok)
	}
}
