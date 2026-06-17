package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Acceptance: specs/model-capability-profiles.feature

// mcpProvLine reproduces the status "Profile  <profile>  (<note>)" wording the
// status spec locks, computed from the same ProfileProvenance the ui renders.
func mcpProvLine(wf *workflow.Workflow, cfg *config.Config) string {
	profile, note := workflow.ProfileProvenance(wf, cfg)
	return "Profile  " + profile + "  (" + note + ")"
}

func mcpAssertLine(t *testing.T, wf *workflow.Workflow, cfg *config.Config, want string) {
	t.Helper()
	if got := mcpProvLine(wf, cfg); got != want {
		t.Fatalf("status line = %q, want %q", got, want)
	}
}

// Scenario: Status shows the profile came from a frontier driver model
func TestMCP_StatusFrontierProvenance(t *testing.T) {
	wf := &workflow.Workflow{DriverModel: "claude-opus-4-7"}
	mcpAssertLine(t, wf, &config.Config{}, "Profile  outcome  (driver: claude-opus-4-7 → frontier)")
}

// Scenario: Status shows strict default for an unknown driver model
func TestMCP_StatusUnknownDriverProvenance(t *testing.T) {
	wf := &workflow.Workflow{DriverModel: "some/unknown-local-model"}
	mcpAssertLine(t, wf, &config.Config{},
		"Profile  strict  (driver: some/unknown-local-model → no capability, default strict)")
}

// Scenario: Status shows the global provenance when an explicit global profile wins
func TestMCP_StatusGlobalProvenance(t *testing.T) {
	wf := &workflow.Workflow{DriverModel: "claude-opus-4-7"}
	mcpAssertLine(t, wf, mcpCfg(config.ProfileGuided, nil, nil), "Profile  guided  (global)")
}

// Scenario: Status shows the per-feature flag provenance when --profile was passed
func TestMCP_StatusFlagProvenance(t *testing.T) {
	wf := &workflow.Workflow{EnforcementProfile: config.ProfileOutcome}
	mcpAssertLine(t, wf, &config.Config{}, "Profile  outcome  (--profile)")
}

// Scenario: Status shows the strict default provenance for a zero-config feature
func TestMCP_StatusDefaultProvenance(t *testing.T) {
	mcpAssertLine(t, &workflow.Workflow{}, &config.Config{}, "Profile  strict  (default)")
}
