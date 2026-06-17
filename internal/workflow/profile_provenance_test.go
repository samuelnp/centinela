package workflow

import (
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// ProfileProvenance returns the active profile and an exact source note for each
// of the five precedence outcomes. The arrow is the Unicode → (U+2192).
func TestProfileProvenance(t *testing.T) {
	cases := []struct {
		name        string
		wf          *Workflow
		cfg         *config.Config
		wantProfile string
		wantNote    string
	}{
		{
			"tier 1 --profile",
			&Workflow{EnforcementProfile: config.ProfileOutcome},
			&config.Config{},
			config.ProfileOutcome, "--profile",
		},
		{
			"tier 2 global",
			&Workflow{DriverModel: "claude-opus-4-7"},
			cfgGlobal(config.ProfileGuided),
			config.ProfileGuided, "global",
		},
		{
			"tier 3 driver hit",
			&Workflow{DriverModel: "claude-opus-4-7"},
			&config.Config{},
			config.ProfileOutcome, "driver: claude-opus-4-7 → frontier",
		},
		{
			"tier 3 driver miss",
			&Workflow{DriverModel: "some/unknown-local-model"},
			&config.Config{},
			config.ProfileStrict, "driver: some/unknown-local-model → no capability, default strict",
		},
		{
			"tier 4 default",
			&Workflow{},
			&config.Config{},
			config.ProfileStrict, "default",
		},
		{
			"nil cfg falls to default",
			&Workflow{DriverModel: "claude-opus-4-7"},
			nil,
			config.ProfileStrict, "default",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			profile, note := ProfileProvenance(tc.wf, tc.cfg)
			if profile != tc.wantProfile || note != tc.wantNote {
				t.Fatalf("got (%q,%q), want (%q,%q)", profile, note, tc.wantProfile, tc.wantNote)
			}
		})
	}
}
