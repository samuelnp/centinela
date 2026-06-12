package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

func epBuildBinary(t *testing.T) string {
	t.Helper()
	o, _ := os.Getwd()
	repo := filepath.Clean(filepath.Join(o, "..", ".."))
	bin := filepath.Join(t.TempDir(), "centinela-ep")
	build := exec.Command("go", "build", "-o", bin, "./cmd/centinela")
	build.Dir = repo
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

// Acceptance: specs/enforcement-profiles.feature

// Scenario: Gates and claim verification run under every profile
func TestEP_GatesRunUnderEveryProfile(t *testing.T) {
	bin := epBuildBinary(t)
	for _, profile := range []string{config.ProfileStrict, config.ProfileGuided, config.ProfileOutcome} {
		dir := t.TempDir()
		toml := "[workflow]\nenforcement_profile=\"" + profile + "\"\n" +
			"[gates]\nfile_size = false\n[validate]\ncommands = [\"exit 1\"]\n"
		os.WriteFile(filepath.Join(dir, config.Filename), []byte(toml), 0644) //nolint:errcheck
		cmd := exec.Command(bin, "validate")
		cmd.Dir = dir
		if err := cmd.Run(); err == nil {
			t.Fatalf("profile %q: failing validate command MUST block completion", profile)
		}
	}
}
