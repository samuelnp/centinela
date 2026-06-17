package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/gitdiff"
)

// Acceptance spec: specs/precommit-and-pr-gate.feature
//
// Integration coverage runs a real temp git repo end-to-end: stage an oversized
// .go file and drive the staged-gate path (gitdiff.Default.ChangedFilesStaged +
// gates.RunWithFilter), asserting the staged file flags G1 while an unstaged
// oversized file is excluded from the staged set. (oversized/mustWrite are
// shared with audit_baseline_ratchet_integration_test.go.)

func gitDo(t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestPrecommit_StagedOversizedFlagsG1_UnstagedExcluded(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git-dependent integration test skipped on windows")
	}
	dir := t.TempDir()
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	gitDo(t, dir, "init", "-q", "-b", "main")
	gitDo(t, dir, "config", "user.email", "qa@centinela.dev")
	gitDo(t, dir, "config", "user.name", "QA")

	mustWrite(t, dir, "internal/oversized.go", oversized(40))
	mustWrite(t, dir, "internal/unstaged.go", oversized(40))
	gitDo(t, dir, "add", "internal/oversized.go")

	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	set, summary, err := gitdiff.Default.ChangedFilesStaged()
	if err != nil {
		t.Fatalf("ChangedFilesStaged: %v", err)
	}
	if summary.Degrade != "" {
		t.Fatalf("real repo must not degrade: %q", summary.Degrade)
	}
	if !set.Contains("internal/oversized.go") {
		t.Fatalf("staged file must be in the set: %v", set.Paths())
	}
	if set.Contains("internal/unstaged.go") {
		t.Fatalf("unstaged file must NOT be in the staged set: %v", set.Paths())
	}

	cfg := &config.Config{}
	cfg.Gates.FileSizeEnabled = true
	if gates.AllPassed(gates.RunWithFilter(cfg, set)) {
		t.Fatal("staged oversized file must flag a fail-severity gate")
	}
}
