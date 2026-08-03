package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// ProfileProvenance returns the active profile and an exact source note for each
// of the five precedence outcomes. The arrow is the Unicode → (U+2192).
func TestProfileProvenance(t *testing.T) {
	// cfg is a factory so tail cases can use a config a real Load produced.
	cases := []struct {
		name        string
		wf          *Workflow
		cfg         func(*testing.T) *config.Config
		wantProfile string
		wantNote    string
	}{
		{
			"tier 1 --profile",
			&Workflow{EnforcementProfile: config.ProfileOutcome},
			loadedCfg,
			config.ProfileOutcome, "--profile",
		},
		{
			"tier 2 global",
			&Workflow{DriverModel: "claude-opus-4-7"},
			func(*testing.T) *config.Config { return cfgGlobal(config.ProfileGuided) },
			config.ProfileGuided, "global",
		},
		{
			"tier 3 driver hit",
			&Workflow{DriverModel: "claude-opus-4-7"},
			loadedCfg,
			config.ProfileOutcome, "driver: claude-opus-4-7 → frontier",
		},
		{
			"tier 3 driver miss",
			&Workflow{DriverModel: "some/unknown-local-model"},
			loadedCfg,
			config.ProfileStrict, "driver: some/unknown-local-model → no capability, default strict",
		},
		{
			"tier 4 pinned contract → guided",
			&Workflow{ProfileContract: ProfileContractGuidedDefault},
			loadedCfg,
			config.ProfileGuided, "default (guided)",
		},
		{
			"tier 4 legacy workflow stays strict",
			&Workflow{},
			loadedCfg,
			config.ProfileStrict, "default (strict, legacy workflow)",
		},
		{
			"nil cfg falls to the legacy default",
			&Workflow{DriverModel: "claude-opus-4-7"},
			func(*testing.T) *config.Config { return nil },
			config.ProfileStrict, "default (strict, legacy workflow)",
		},
		{
			"capability guided is NOT conflated with default guided",
			&Workflow{DriverModel: "claude-sonnet-4-6", ProfileContract: ProfileContractGuidedDefault},
			loadedCfg,
			config.ProfileGuided, "driver: claude-sonnet-4-6 → capable",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			profile, note := ProfileProvenance(tc.wf, tc.cfg(t))
			if profile != tc.wantProfile || note != tc.wantNote {
				t.Fatalf("got (%q,%q), want (%q,%q)", profile, note, tc.wantProfile, tc.wantNote)
			}
		})
	}
}
