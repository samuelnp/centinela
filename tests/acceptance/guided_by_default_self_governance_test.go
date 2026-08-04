// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// Scenario: Centinela's own repository pins its profile explicitly
//
// The plan's Slice 5 guard: if this repo's centinela.toml ever drops the
// explicit enforcement_profile = "strict" pin, this test fails rather than
// letting Centinela's own dogfooding silently inherit the new guided default.
func TestGBD_RepoPinsEnforcementProfileExplicitly(t *testing.T) {
	t.Chdir(repoRoot(t))
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("this repo's centinela.toml must parse: %v", err)
	}
	if cfg.Workflow.RawEnforcementProfile == "" {
		t.Fatal("this repo's centinela.toml must pin enforcement_profile explicitly — " +
			"the shipped default is now guided, and Centinela's own dogfooding must not " +
			"inherit it silently")
	}
	if cfg.Workflow.RawEnforcementProfile != config.ProfileStrict {
		t.Fatalf("this repo's pinned profile must be %q, got %q",
			config.ProfileStrict, cfg.Workflow.RawEnforcementProfile)
	}
}
