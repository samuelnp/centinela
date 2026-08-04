package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// TestProfileResolutionSurfacesAgree walks every tier — including the case the
// other profile tests structurally miss: a workflow that carries the guided
// contract pin AND declares a driver model with no capability class.
//
// EffectiveProfile feeds `centinela verdict` (and the MCP governance payload)
// while ProfileProvenance feeds `centinela status`. If they disagree, two
// shipped surfaces report different profiles for one state file and at least one
// of them is lying. The unmapped-driver row is the regression guard: the tail
// used to swallow it and answer guided while status answered strict.
func TestProfileResolutionSurfacesAgree(t *testing.T) {
	pin := func(wf *Workflow) *Workflow { wf.ProfileContract = ProfileContractGuidedDefault; return wf }
	// cfg is built INSIDE each subtest: a tail-reaching case needs a config a
	// real Load produced, which loadedCfg supplies (it chdirs into a temp dir).
	cases := []struct {
		name string
		wf   *Workflow
		cfg  func(*testing.T) *config.Config
		want string
	}{
		{"pinned + unmapped driver", pin(&Workflow{DriverModel: "totally-unknown-model"}),
			loadedCfg, config.ProfileStrict},
		{"unpinned + unmapped driver", &Workflow{DriverModel: "totally-unknown-model"},
			loadedCfg, config.ProfileStrict},
		{"pinned + capable driver", pin(&Workflow{DriverModel: "sonnet"}), loadedCfg, config.ProfileGuided},
		{"pinned + limited driver", pin(&Workflow{DriverModel: "haiku"}), loadedCfg, config.ProfileStrict},
		{"pinned + frontier driver", pin(&Workflow{DriverModel: "opus"}), loadedCfg, config.ProfileOutcome},
		{"pinned, no driver", pin(&Workflow{}), loadedCfg, config.ProfileGuided},
		{"legacy, no driver", &Workflow{}, loadedCfg, config.ProfileStrict},
		{"explicit global outranks an unmapped driver",
			pin(&Workflow{DriverModel: "totally-unknown-model"}),
			func(*testing.T) *config.Config { return cfgWithProfile(config.ProfileOutcome) },
			config.ProfileOutcome},
		{"per-feature pin outranks an unmapped driver",
			pin(&Workflow{DriverModel: "totally-unknown-model", EnforcementProfile: config.ProfileGuided}),
			loadedCfg, config.ProfileGuided},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.cfg(t)
			effective := EffectiveProfile(tc.wf, cfg)
			reported, note := ProfileProvenance(tc.wf, cfg)
			if effective != tc.want {
				t.Errorf("EffectiveProfile = %q, want %q", effective, tc.want)
			}
			if effective != reported {
				t.Fatalf("surfaces disagree: verdict says %q, status says %q (%s)",
					effective, reported, note)
			}
		})
	}
}

// TestUnmappedDriverNeverInheritsTheGuidedTail states the rule on its own so a
// future edit that reorders the tiers cannot quietly restore the fall-through.
func TestUnmappedDriverNeverInheritsTheGuidedTail(t *testing.T) {
	loaded := loadedCfg(t)
	wf := &Workflow{DriverModel: "who/knows", ProfileContract: ProfileContractGuidedDefault}
	if got := EffectiveProfile(wf, loaded); got != config.ProfileStrict {
		t.Fatalf("a declared-but-unmapped driver must keep maximum scaffolding, got %q", got)
	}
	// Tier 3 is terminal only when a driver is DECLARED: with none, the pinned
	// contract still reaches the guided tail.
	if got := EffectiveProfile(&Workflow{ProfileContract: ProfileContractGuidedDefault},
		loaded); got != config.ProfileGuided {
		t.Fatalf("the tail must still fire when no driver is declared, got %q", got)
	}
	// A nil config is not "no capability tier to miss" — it is a config nobody
	// loaded, and that resolves strict.
	if got := EffectiveProfile(wf, nil); got != config.ProfileStrict {
		t.Fatalf("a nil config must fail closed, got %q", got)
	}
}
