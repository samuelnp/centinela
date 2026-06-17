package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func cfgCap(modelClass map[string]string) *config.Config {
	c := &config.Config{}
	c.Orchestration.Capabilities = modelClass
	return c
}

// EffectiveProfile tier 3: the capability default derived from wf.DriverModel,
// plus the tier-3 miss (driver set, no class → strict) and the zero default.
func TestEffectiveProfile_CapabilityTier(t *testing.T) {
	cases := []struct {
		name   string
		driver string
		cfg    *config.Config
		want   string
	}{
		{"frontier builtin → outcome", "claude-opus-4-7", &config.Config{}, config.ProfileOutcome},
		{"declared limited → strict", "local/weak", cfgCap(map[string]string{"local/weak": "limited"}), config.ProfileStrict},
		{"declared capable → guided", "local/cap", cfgCap(map[string]string{"local/cap": "capable"}), config.ProfileGuided},
		{"driver set but no class → strict", "some/unknown", &config.Config{}, config.ProfileStrict},
		{"no driver, zero cfg → strict", "", &config.Config{}, config.ProfileStrict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wf := &Workflow{DriverModel: tc.driver}
			if got := EffectiveProfile(wf, tc.cfg); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// The load-bearing back-compat guarantee: zero workflow + zero config → strict.
func TestEffectiveProfile_ZeroConfigStrict(t *testing.T) {
	if got := EffectiveProfile(&Workflow{}, &config.Config{}); got != config.ProfileStrict {
		t.Fatalf("zero-config must resolve strict, got %q", got)
	}
}

// A pinned --profile and an explicit global both win over a frontier driver.
func TestEffectiveProfile_HigherTiersBeatCapability(t *testing.T) {
	cfg := &config.Config{}
	cfg.Workflow.EnforcementProfile = config.ProfileGuided
	cfg.Workflow.RawEnforcementProfile = config.ProfileGuided
	wf := &Workflow{DriverModel: "claude-opus-4-7"}
	if got := EffectiveProfile(wf, cfg); got != config.ProfileGuided {
		t.Fatalf("explicit global must beat frontier capability, got %q", got)
	}
	pinned := &Workflow{EnforcementProfile: config.ProfileStrict, DriverModel: "claude-opus-4-7"}
	if got := EffectiveProfile(pinned, cfg); got != config.ProfileStrict {
		t.Fatalf("pinned --profile must beat everything, got %q", got)
	}
}
