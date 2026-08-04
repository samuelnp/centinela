// Acceptance: specs/guided-by-default.feature
package acceptance_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// Scenario: Gate failures block completion identically under strict and guided
//
// Reuses epBuildBinary (enforcement_profiles_invariant_test.go): the same
// binary and the same failing `validate.commands` entry, scoped to exactly
// the two profiles this spec is about.
func TestGBD_GateFailureBlocksBothProfiles(t *testing.T) {
	bin := epBuildBinary(t)
	for _, profile := range []string{config.ProfileStrict, config.ProfileGuided} {
		dir := t.TempDir()
		toml := "[workflow]\nenforcement_profile=\"" + profile + "\"\n" +
			"[gates]\nfile_size = false\n[validate]\ncommands = [\"exit 1\"]\n"
		mustWrite(t, filepath.Join(dir, config.Filename), toml)
		cmd := exec.Command(bin, "validate")
		cmd.Dir = dir
		if err := cmd.Run(); err == nil {
			t.Fatalf("profile %q: a failing validate command must block completion identically", profile)
		}
	}
}
