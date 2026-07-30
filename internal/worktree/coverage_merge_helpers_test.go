package worktree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
)

// parseConflictedPaths returns nil when git cannot run (e.g. not a repo).
func TestParseConflictedPaths_NonRepoReturnsNil(t *testing.T) {
	if got := parseConflictedPaths(t.TempDir()); got != nil {
		t.Fatalf("parseConflictedPaths outside a repo = %v, want nil", got)
	}
}

// Replaces TestScenariosConflicts_SkipsEmptyFields, dropped with the
// Given-only pairing removed by the spec-conflict-false-positives hotfix.
// indexByKey keeps the first record when a file repeats a scenario name, so a
// duplicate cannot manufacture extra pairings.
func TestIndexByKey_FirstRecordWins(t *testing.T) {
	recs := []scenarioRecord{
		{Owner: "a", File: "f.feature", Scenario: "s", Given: "g", Then: "1"},
		{Owner: "a", File: "f.feature", Scenario: "s", Given: "g", Then: "2"},
	}
	idx := indexByKey(recs)
	if len(idx) != 1 {
		t.Fatalf("indexByKey = %d entries, want 1", len(idx))
	}
	if got := idx["f.feature\x00s"].Then; got != "1" {
		t.Fatalf("first record should win, got Then=%q", got)
	}
}

// readSpecsFrom continues past a .feature entry it cannot read (a dangling
// symlink: listed by ReadDir, not a dir, but ReadFile fails).
func TestReadSpecsFrom_UnreadableEntrySkipped(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "broken.feature")
	if err := os.Symlink(filepath.Join(dir, "missing"), link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}
	if recs := readSpecsFrom(dir, "main"); len(recs) != 0 {
		t.Fatalf("readSpecsFrom should skip unreadable entries, got %v", recs)
	}
}

// MaybeProvision surfaces the chdir failure when Create reports success but
// the worktree path was never materialised (git stubbed to a no-op).
func TestMaybeProvision_ChdirFailureSurfaced(t *testing.T) {
	old := gitRunner
	defer func() { gitRunner = old }()
	gitRunner = func(string, ...string) ([]byte, error) { return nil, nil }

	repo := t.TempDir()
	cfg := &config.Config{}
	cfg.Workflow.UseWorktrees = true
	if _, err := MaybeProvision(repo, "feat", cfg); err == nil {
		t.Fatal("MaybeProvision must error when it cannot chdir into the worktree")
	}
}
