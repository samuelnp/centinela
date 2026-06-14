package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitAvailable(t *testing.T) {
	stubGit(t, okGit(""))
	if !gitAvailable("x") {
		t.Fatal("rev-parse=true must report available")
	}
	stubGit(t, func(string, ...string) ([]byte, error) { return nil, errStub })
	if gitAvailable("x") {
		t.Fatal("error must report unavailable")
	}
}

func TestListWorktreesSorted(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"b", "a"} {
		if err := os.MkdirAll(filepath.Join(dir, ".worktrees", n), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// a non-dir file is ignored.
	_ = os.WriteFile(filepath.Join(dir, ".worktrees", "f.txt"), []byte("x"), 0o644)
	got := listWorktrees(dir)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("expected sorted [a b], got %v", got)
	}
	if listWorktrees(filepath.Join(dir, "nope")) != nil {
		t.Fatal("missing .worktrees must return nil")
	}
}

func TestWorkflowDone(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, ".workflow")
	_ = os.MkdirAll(wf, 0o755)
	_ = os.WriteFile(filepath.Join(wf, "done.json"), []byte(`{"currentStep": "done"}`), 0o644)
	_ = os.WriteFile(filepath.Join(wf, "wip.json"), []byte(`{"currentStep": "code"}`), 0o644)
	if !workflowDone(dir, "done") {
		t.Fatal("done step must report done")
	}
	if workflowDone(dir, "wip") {
		t.Fatal("code step must not report done")
	}
	if workflowDone(dir, "missing") {
		t.Fatal("missing file must not report done")
	}
}

func TestBranchMergedMissingBranchIsMerged(t *testing.T) {
	stubGit(t, func(repo string, args ...string) ([]byte, error) { return nil, errStub })
	if !branchMerged("x", "gone") {
		t.Fatal("missing branch counts as merged")
	}
}

func TestBranchMergedListError(t *testing.T) {
	stubGit(t, func(repo string, args ...string) ([]byte, error) {
		if args[0] == "rev-parse" {
			return []byte("ok"), nil
		}
		return nil, errStub
	})
	if branchMerged("x", "feat") {
		t.Fatal("branch list error must report not merged")
	}
}

func TestJoinCommands(t *testing.T) {
	if joinCommands(nil) != "" {
		t.Fatal("empty must be blank")
	}
	if got := joinCommands([]string{"a", "b"}); got != "a && b" {
		t.Fatalf("join: %q", got)
	}
}
