package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoadmapCheckCorruptLoadError(t *testing.T) {
	repoFixture(t)
	writeFile(t, ".workflow/roadmap.json", "{ not json")
	d := roadmapCheck{}.Run(Context{})
	if d.Status != Error || !strings.Contains(d.Message, "cannot load") {
		t.Fatalf("corrupt roadmap must Error, got %v %q", d.Status, d.Message)
	}
}

func TestConfigCheckUnknownKeysWarnViaRun(t *testing.T) {
	repoFixture(t)
	writeFile(t, "centinela.toml", "bogus = 1\n")
	d := configCheck{}.Run(Context{Config: configWithTimeout(240)})
	if d.Status != Warn {
		t.Fatalf("unknown key must Warn, got %v", d.Status)
	}
	if !strings.Contains(strings.Join(d.Details, " "), "bogus") {
		t.Fatalf("unknown key must be named: %v", d.Details)
	}
}

func TestVersionCheckEmptyMakefileOK(t *testing.T) {
	dir := repoFixture(t) // no Makefile
	stubVersion(t, func() (string, error) { return "0.9.9\n", nil })
	d := versionCheck{}.Run(Context{Root: dir})
	if d.Status != OK {
		t.Fatalf("no Makefile version must be OK (cannot compare), got %v %q", d.Status, d.Message)
	}
}

func TestOrphanedWorkflowsSkipsLiveWorktree(t *testing.T) {
	dir := repoFixture(t)
	seedWorkflow(t, "feat")
	if err := os.MkdirAll(filepath.Join(dir, ".worktrees", "feat"), 0o755); err != nil {
		t.Fatal(err)
	}
	stubGit(t, func(repo string, args ...string) ([]byte, error) {
		if args[0] == "rev-parse" && len(args) > 1 && args[1] == "--is-inside-work-tree" {
			return []byte("true\n"), nil
		}
		return nil, errStub
	})
	if got := orphanedWorkflows(dir); len(got) != 0 {
		t.Fatalf("workflow with a live worktree is not orphaned, got %v", got)
	}
}

func TestRepairEvidenceErrorPropagates(t *testing.T) {
	repoFixture(t)
	writeFile(t, ".workflow/feat-qa-senior.json.tmp", "{}")
	// make .workflow read+exec but not writable so os.Remove of the tmp fails.
	if err := os.Chmod(".workflow", 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(".workflow", 0o755) })
	if err := repairEvidence(); err == nil {
		t.Fatal("repairEvidence must surface a remove failure")
	}
}
