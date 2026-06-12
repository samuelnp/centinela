package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func cfgGlobal(profile string) *config.Config {
	c := &config.Config{}
	c.Workflow.EnforcementProfile = profile
	c.Workflow.RawEnforcementProfile = profile
	return c
}

// ResolveStart: --profile pins and sets the effective mode; an explicit global is
// honored for the effective mode but deliberately NOT pinned (the correctness fix);
// a capability default is unpinned; a driver miss and zero both fall to strict.
func TestResolveStart(t *testing.T) {
	cases := []struct {
		name          string
		flagProfile   string
		flagModel     string
		cfg           *config.Config
		wantPinned    string
		wantEffective string
		wantDriver    string
	}{
		{"--profile pins", "outcome", "", &config.Config{}, config.ProfileOutcome, config.ProfileOutcome, ""},
		{"explicit global not pinned", "", "", cfgGlobal(config.ProfileGuided), "", config.ProfileGuided, ""},
		{"capability default unpinned", "", "claude-opus-4-7", &config.Config{}, "", config.ProfileOutcome, "claude-opus-4-7"},
		{"driver miss → strict", "", "some/unknown", &config.Config{}, "", config.ProfileStrict, "some/unknown"},
		{"zero → strict", "", "", &config.Config{}, "", config.ProfileStrict, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := ResolveStart(tc.flagProfile, tc.flagModel, tc.cfg)
			if d.PinnedProfile != tc.wantPinned {
				t.Fatalf("PinnedProfile = %q, want %q", d.PinnedProfile, tc.wantPinned)
			}
			if d.EffectiveProfile != tc.wantEffective {
				t.Fatalf("EffectiveProfile = %q, want %q", d.EffectiveProfile, tc.wantEffective)
			}
			if d.DriverModel != tc.wantDriver {
				t.Fatalf("DriverModel = %q, want %q", d.DriverModel, tc.wantDriver)
			}
		})
	}
}

// Driver-model resolution (flag → env → config) is surfaced through ResolveStart.
func TestResolveStart_DriverResolution(t *testing.T) {
	t.Setenv("CENTINELA_MODEL", "env-model")
	c := &config.Config{}
	c.Orchestration.DriverModel = "config-model"
	if d := ResolveStart("", "flag-model", c); d.DriverModel != "flag-model" {
		t.Fatalf("flag must win, got %q", d.DriverModel)
	}
	if d := ResolveStart("", "", c); d.DriverModel != "env-model" {
		t.Fatalf("env must win over config, got %q", d.DriverModel)
	}
}
