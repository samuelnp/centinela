package main

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// THE INVARIANT TEST (spec scenario "Gates and claim verification run under
// every profile"). For EACH profile, a validate-step project whose validate
// command exits non-zero must be BLOCKED — the verdict does not depend on the
// configured profile. executeValidation() is exactly the gate complete() runs at
// the validate step; profiles never branch around it.
func TestExecuteValidation_BlockedUnderEveryProfile(t *testing.T) {
	for _, profile := range []string{config.ProfileStrict, config.ProfileGuided, config.ProfileOutcome} {
		t.Run(profile, func(t *testing.T) {
			t.Chdir(t.TempDir())
			// file_size off isolates the failure to the failing command; the
			// enforcement_profile is the only variable across the three runs.
			toml := "[workflow]\nenforcement_profile=\"" + profile + "\"\n" +
				"[gates]\nfile_size = false\n" +
				"[validate]\ncommands = [\"exit 1\"]\n"
			os.WriteFile(config.Filename, []byte(toml), 0644) //nolint:errcheck

			if err := executeValidation(); err == nil {
				t.Fatalf("profile %q: failing validate command MUST block, got nil error", profile)
			}
		})
	}
}

// Control: with a passing command, validation succeeds under every profile too —
// proving the block above is caused by the failing command, not the profile.
func TestExecuteValidation_PassesUnderEveryProfileWhenClean(t *testing.T) {
	for _, profile := range []string{config.ProfileStrict, config.ProfileGuided, config.ProfileOutcome} {
		t.Run(profile, func(t *testing.T) {
			t.Chdir(t.TempDir())
			toml := "[workflow]\nenforcement_profile=\"" + profile + "\"\n" +
				"[gates]\nfile_size = false\n" +
				"[validate]\ncommands = [\"exit 0\"]\n"
			os.WriteFile(config.Filename, []byte(toml), 0644) //nolint:errcheck
			if err := executeValidation(); err != nil {
				t.Fatalf("profile %q: clean validate must pass, got %v", profile, err)
			}
		})
	}
}
