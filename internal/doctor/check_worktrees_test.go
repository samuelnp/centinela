package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func okGit(out string) func(string, ...string) ([]byte, error) {
	return func(repo string, args ...string) ([]byte, error) {
		if len(args) > 1 && args[0] == "rev-parse" && args[1] == "--is-inside-work-tree" {
			return []byte("true\n"), nil
		}
		return []byte(out), nil
	}
}

func TestWorktreesCheckNoGitWarns(t *testing.T) {
	dir := repoFixture(t)
	stubGit(t, func(string, ...string) ([]byte, error) { return nil, os.ErrNotExist })
	d := worktreesCheck{}.Run(Context{Root: dir})
	if d.Status != Warn || !strings.Contains(d.Message, "git context") {
		t.Fatalf("no git must Warn, got %v %q", d.Status, d.Message)
	}
}

func TestWorktreesCheckNoneOK(t *testing.T) {
	dir := repoFixture(t)
	stubGit(t, okGit(""))
	d := worktreesCheck{}.Run(Context{Root: dir})
	if d.Status != OK {
		t.Fatalf("no worktrees must be OK, got %v", d.Status)
	}
}

func TestWorktreesCheckAbandonedReportsCommand(t *testing.T) {
	dir := repoFixture(t)
	if err := os.MkdirAll(filepath.Join(dir, ".worktrees", "gone"), 0o755); err != nil {
		t.Fatal(err)
	}
	// branch missing => merged => abandoned.
	stubGit(t, func(repo string, args ...string) ([]byte, error) {
		if args[0] == "rev-parse" && args[1] == "--is-inside-work-tree" {
			return []byte("true\n"), nil
		}
		if args[0] == "rev-parse" {
			return nil, os.ErrNotExist
		}
		return []byte(""), nil
	})
	d := worktreesCheck{}.Run(Context{Root: dir})
	if d.Status != Error {
		t.Fatalf("abandoned worktree must Error, got %v", d.Status)
	}
	if d.Repair == nil || d.Repair.Apply != nil {
		t.Fatal("worktrees repair must be report-only (Apply nil)")
	}
	if !strings.Contains(d.Repair.Command, "git worktree remove") {
		t.Fatalf("command must be the remove command: %q", d.Repair.Command)
	}
}

func TestBranchMergedHelper(t *testing.T) {
	dir := t.TempDir()
	stubGit(t, func(repo string, args ...string) ([]byte, error) {
		if args[0] == "rev-parse" {
			return []byte("ok"), nil
		}
		return []byte("  feat\n* main\n"), nil
	})
	if !branchMerged(dir, "feat") {
		t.Fatal("feat listed under --merged must be merged")
	}
	if branchMerged(dir, "other") {
		t.Fatal("unlisted branch must not be merged")
	}
}
