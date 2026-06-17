package doctor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/setup"
)

// errStub is a generic error for git-runner stubs simulating a missing ref.
var errStub = errors.New("stub error")

// repoFixture chdirs into a fresh temp repo with a .workflow/ dir and returns
// its path. t.Chdir auto-restores the working directory at test end.
func repoFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	t.Chdir(dir)
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// writeFile writes content under the current dir, creating parents.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// seedRoadmap writes a roadmap.json with the given phase name and a matching,
// in-sync ROADMAP.md.
func seedRoadmap(t *testing.T, phaseName string) {
	t.Helper()
	rm := &roadmap.Roadmap{Phases: []roadmap.Phase{{
		Name:     phaseName,
		Features: []roadmap.Feature{{Name: "f", Description: "d"}},
	}}}
	if err := roadmap.Save(rm); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("ROADMAP.md", roadmap.RenderMarkdown(rm), 0o644); err != nil {
		t.Fatal(err)
	}
}

// seedSyncedHooks fully wires .claude + opencode managed files so the hooks
// check observes a no-change (OK) plan.
func seedSyncedHooks(t *testing.T) {
	t.Helper()
	p, err := setup.BuildSyncPlan("both")
	if err != nil {
		t.Fatal(err)
	}
	if err := setup.ApplySync(p); err != nil {
		t.Fatal(err)
	}
}

// configWithTimeout returns a minimal config with the given verify_timeout.
func configWithTimeout(secs int) *config.Config {
	cfg := &config.Config{}
	cfg.Verify.TimeoutSeconds = secs
	return cfg
}

// stubGit installs a gitRunner stub and restores it after the test.
func stubGit(t *testing.T, fn func(repo string, args ...string) ([]byte, error)) {
	t.Helper()
	orig := gitRunner
	gitRunner = fn
	t.Cleanup(func() { gitRunner = orig })
}

// stubVersion installs a versionRunner stub and restores it after the test.
func stubVersion(t *testing.T, fn func() (string, error)) {
	t.Helper()
	orig := versionRunner
	versionRunner = fn
	t.Cleanup(func() { versionRunner = orig })
}
