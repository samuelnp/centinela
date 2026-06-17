package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Acceptance: specs/model-capability-profiles.feature

func mcpCfg(global string, caps, profiles map[string]string) *config.Config {
	c := &config.Config{}
	c.Workflow.EnforcementProfile = global
	c.Workflow.RawEnforcementProfile = global // mirrors config.Load's raw capture
	c.Orchestration.Capabilities = caps
	c.Orchestration.CapabilityProfiles = profiles
	return c
}

func mcpEffective(t *testing.T, wf *workflow.Workflow, cfg *config.Config, want string) {
	t.Helper()
	if got := workflow.EffectiveProfile(wf, cfg); got != want {
		t.Fatalf("effective profile = %q, want %q", got, want)
	}
}

// Scenario: Zero config resolves to strict byte-identically
func TestMCP_ZeroConfigStrict(t *testing.T) {
	mcpEffective(t, &workflow.Workflow{}, &config.Config{}, config.ProfileStrict)
}

// Scenario: Frontier built-in driver model defaults to outcome
func TestMCP_FrontierDefaultsOutcome(t *testing.T) {
	wf := &workflow.Workflow{DriverModel: "claude-opus-4-7"}
	mcpEffective(t, wf, &config.Config{}, config.ProfileOutcome)
}

// Scenario: Capable local driver model declared in config defaults to guided
func TestMCP_CapableLocalDefaultsGuided(t *testing.T) {
	wf := &workflow.Workflow{DriverModel: "local/some-capable"}
	cfg := mcpCfg("", map[string]string{"local/some-capable": "capable"}, nil)
	mcpEffective(t, wf, cfg, config.ProfileGuided)
}

// Scenario: Limited local driver model declared in config defaults to strict
func TestMCP_LimitedLocalDefaultsStrict(t *testing.T) {
	wf := &workflow.Workflow{DriverModel: "local/weak-model"}
	cfg := mcpCfg("", map[string]string{"local/weak-model": "limited"}, nil)
	mcpEffective(t, wf, cfg, config.ProfileStrict)
}

// Scenario: Unknown driver model with no capability falls back to strict
func TestMCP_UnknownDriverFallsToStrict(t *testing.T) {
	wf := &workflow.Workflow{DriverModel: "some/unknown-local-model"}
	mcpEffective(t, wf, &config.Config{}, config.ProfileStrict)
}

// Scenario: Capability profiles override remaps a class to a different profile
func TestMCP_CapabilityProfilesOverride(t *testing.T) {
	wf := &workflow.Workflow{DriverModel: "claude-opus-4-7"}
	cfg := mcpCfg("", nil, map[string]string{"frontier": "guided"})
	mcpEffective(t, wf, cfg, config.ProfileGuided)
}

// Scenario: Explicit global enforcement_profile beats the capability default
func TestMCP_ExplicitGlobalBeatsCapability(t *testing.T) {
	wf := &workflow.Workflow{DriverModel: "claude-opus-4-7"}
	mcpEffective(t, wf, mcpCfg(config.ProfileGuided, nil, nil), config.ProfileGuided)
}

// Scenario: Per-feature profile flag beats the capability default
func TestMCP_FlagBeatsCapability(t *testing.T) {
	wf := &workflow.Workflow{EnforcementProfile: config.ProfileOutcome, DriverModel: "claude-haiku-4-5-20251001"}
	mcpEffective(t, wf, mcpCfg(config.ProfileStrict, nil, nil), config.ProfileOutcome)
}
