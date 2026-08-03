package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// TestMCPRulesReportsWorkflowScopeNotProjectScope: read_rules reports archetype
// from the ACTIVE workflow, so its profile must come from the same scope. A
// project-scoped profile beside a workflow-scoped archetype mixes two scopes in
// one payload, and it disagrees with status and verdict on every legacy
// workflow — whose pin keeps it strict while the project default is guided.
func TestMCPRulesReportsWorkflowScopeNotProjectScope(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("CENTINELA_MODEL", "")
	cfg, err := config.LoadForProfile()
	if err != nil {
		t.Fatalf("LoadForProfile: %v", err)
	}
	legacy := &workflow.Workflow{Feature: "f", CurrentStep: "code"} // no contract pin
	if got := rulesProfile(legacy, cfg); got != config.ProfileStrict {
		t.Fatalf("a legacy workflow governs read_rules as strict, got %q", got)
	}
	if got := workflow.EffectiveProfile(legacy, cfg); got != config.ProfileStrict {
		t.Fatalf("verdict must agree, got %q", got)
	}
	// Project scope is the fallback ONLY when there is no workflow to govern.
	if got := rulesProfile(nil, cfg); got != config.ProjectDefaultProfile(cfg) {
		t.Fatalf("with no workflow, read_rules must report the project default, got %q", got)
	}
	if config.ProjectDefaultProfile(cfg) != config.ProfileGuided {
		t.Fatal("fixture check: a zero-config project should default to guided")
	}
}
