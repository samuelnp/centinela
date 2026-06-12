package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Acceptance: specs/model-capability-profiles.feature

func mcpDriverCfg(driver string) *config.Config {
	c := &config.Config{}
	c.Orchestration.DriverModel = driver
	return c
}

// Scenario: Driver model flag overrides env overrides config
func TestMCP_DriverFlagOverridesEnvOverridesConfig(t *testing.T) {
	t.Setenv("CENTINELA_MODEL", "env-model")
	d := workflow.ResolveStart("", "flag-model", mcpDriverCfg("config-model"))
	if d.DriverModel != "flag-model" {
		t.Fatalf("pinned driver = %q, want flag-model", d.DriverModel)
	}
}

// Scenario: Driver model env overrides config when no flag is given
func TestMCP_DriverEnvOverridesConfig(t *testing.T) {
	t.Setenv("CENTINELA_MODEL", "env-model")
	d := workflow.ResolveStart("", "", mcpDriverCfg("config-model"))
	if d.DriverModel != "env-model" {
		t.Fatalf("pinned driver = %q, want env-model", d.DriverModel)
	}
}

// Scenario: Driver model falls back to config when no flag and no env
func TestMCP_DriverFallsToConfig(t *testing.T) {
	t.Setenv("CENTINELA_MODEL", "")
	d := workflow.ResolveStart("", "", mcpDriverCfg("config-model"))
	if d.DriverModel != "config-model" {
		t.Fatalf("pinned driver = %q, want config-model", d.DriverModel)
	}
}

// Scenario: Driver model is empty when nothing is configured
func TestMCP_DriverEmptyWhenUnconfigured(t *testing.T) {
	t.Setenv("CENTINELA_MODEL", "")
	d := workflow.ResolveStart("", "", &config.Config{})
	if d.DriverModel != "" {
		t.Fatalf("pinned driver = %q, want empty", d.DriverModel)
	}
}

// Scenario: An opaque model id with no capability is accepted and pins without error
func TestMCP_OpaqueModelAcceptedAndPins(t *testing.T) {
	t.Setenv("CENTINELA_MODEL", "")
	d := workflow.ResolveStart("", "totally/made-up-model", &config.Config{})
	if d.DriverModel != "totally/made-up-model" {
		t.Fatalf("pinned driver = %q, want totally/made-up-model", d.DriverModel)
	}
	// No error path exists in resolution: an unknown id simply yields no capability
	// tier (strict), which the precedence suite asserts. The pin itself never errors.
	if d.EffectiveProfile != config.ProfileStrict {
		t.Fatalf("opaque model effective = %q, want strict", d.EffectiveProfile)
	}
}
