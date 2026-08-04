package workflow

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/treestate"
)

// freshRepo seeds a real one-commit repository, chdirs into it and stamps a
// gatekeeper report against it. Stubs cannot prove the range semantics of
// `git diff --name-only A..HEAD`; only git can.
func freshRepo(t *testing.T) (string, treestate.Runner) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	t.Chdir(dir)
	realWrite(t, dir, "internal/example/thing.go", "package example\n")
	realWrite(t, dir, "ROADMAP.md", "# Roadmap\n")
	realWrite(t, dir, ".workflow/roadmap.json", "{}\n")
	realGit(t, dir, "init", "-q", "-b", "main")
	realGit(t, dir, "config", "user.email", "t@e.com")
	realGit(t, dir, "config", "user.name", "T")
	realGit(t, dir, "config", "commit.gpgsign", "false")
	realGit(t, dir, "add", "-A")
	realGit(t, dir, "commit", "-qm", "seed")
	run := treestate.NewExecRunner()
	if err := os.MkdirAll(filepath.Join(dir, WorkflowDir), 0o755); err != nil {
		t.Fatal(err)
	}
	wf := New("f")
	wf.ValidateContract = ValidateContractAdversarial
	if err := Save(wf); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(GatekeeperReportPath("f"), []byte(freshReport(t, run)), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, run
}

func realWrite(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func realGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
}

// AC9 against real git: a roadmap-state commit landing after the stamp keeps
// the verification fresh, and a source commit in the same range stales it.
func TestVerificationFreshAcrossARealRoadmapStateCommit(t *testing.T) {
	dir, run := freshRepo(t)
	realWrite(t, dir, ".workflow/roadmap.json", `{"phases":[]}`)
	realWrite(t, dir, "ROADMAP.md", "# Roadmap\n\nchanged\n")
	realGit(t, dir, "commit", "--no-verify", "-q", "-m", "chore(roadmap): defer x",
		"--", ".workflow/roadmap.json", "ROADMAP.md")
	if err := VerificationFresh("f", dir, run); err != nil {
		t.Fatalf("a real roadmap-state commit must not stale the stamp: %v", err)
	}
	realWrite(t, dir, "internal/example/thing.go", "package example // edited\n")
	realGit(t, dir, "commit", "--no-verify", "-q", "-am", "feat: real change")
	if err := VerificationFresh("f", dir, run); err == nil {
		t.Fatal("a source commit in the same range must stale the stamp")
	}
}
