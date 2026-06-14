package acceptance_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/setup"
)

// gitInit turns dir into a git repo on branch main with one seed commit.
func gitInit(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	run("config", "user.email", "qa@centinela.dev")
	run("config", "user.name", "QA")
	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("seed\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".worktrees/\n"), 0o644)
	run("add", ".")
	run("commit", "-q", "-m", "seed")
}

// seedHooks wires all managed hook/config files so the hooks check is OK.
func seedHooks(t *testing.T, dir string) {
	t.Helper()
	withDir(t, dir, func() {
		p, err := setup.BuildSyncPlan("both")
		if err != nil {
			t.Fatal(err)
		}
		if err := setup.ApplySync(p); err != nil {
			t.Fatal(err)
		}
	})
}

// seedRoadmap writes an in-sync roadmap.json + ROADMAP.md with phaseName.
func seedRoadmap(t *testing.T, dir, phaseName string) {
	t.Helper()
	withDir(t, dir, func() {
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
	})
}

// withDir runs fn with the process CWD temporarily set to dir.
func withDir(t *testing.T, dir string, fn func()) {
	t.Helper()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()
	fn()
}

// writeFile writes content to a path under dir, creating parents.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
