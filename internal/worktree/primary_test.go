package worktree

import (
	"testing"
)

// stubGit swaps the package git seam for the duration of a test.
func stubGit(t *testing.T, fn func(repo string, args ...string) ([]byte, error)) {
	t.Helper()
	old := gitRunner
	t.Cleanup(func() { gitRunner = old })
	gitRunner = fn
}

// porcelainStub makes every git call return out verbatim.
func porcelainStub(t *testing.T, out string) {
	t.Helper()
	stubGit(t, func(string, ...string) ([]byte, error) { return []byte(out), nil })
}

// The first `worktree <path>` block of a multi-worktree listing is the
// primary tree — never the feature worktree the command may run from.
func TestPrimaryTree_MultiWorktree_FirstEntryWins(t *testing.T) {
	porcelainStub(t, "worktree /repo\nHEAD abc\nbranch refs/heads/main\n\n"+
		"worktree /repo/.worktrees/feat\nHEAD def\nbranch refs/heads/feat\n\n")
	got, err := PrimaryTree(".")
	if err != nil {
		t.Fatalf("PrimaryTree: %v", err)
	}
	if got != "/repo" {
		t.Fatalf("PrimaryTree = %q, want /repo", got)
	}
}

// A single-tree listing without a trailing blank line still resolves.
func TestPrimaryTree_SingleTree_NoTrailingBlank(t *testing.T) {
	porcelainStub(t, "worktree /solo\nHEAD abc\nbranch refs/heads/main")
	got, err := PrimaryTree(".")
	if err != nil {
		t.Fatalf("PrimaryTree: %v", err)
	}
	if got != "/solo" {
		t.Fatalf("PrimaryTree = %q, want /solo", got)
	}
}

// Paths may contain spaces: the value is everything after the first space.
// CRLF endings and trailing blank lines are tolerated.
func TestPrimaryTree_PathWithSpacesAndCRLF(t *testing.T) {
	porcelainStub(t, "worktree /Users/dev/My Projects/repo\r\nHEAD abc\r\n\r\n\r\n")
	got, err := PrimaryTree(".")
	if err != nil {
		t.Fatalf("PrimaryTree: %v", err)
	}
	if got != "/Users/dev/My Projects/repo" {
		t.Fatalf("PrimaryTree = %q, want the space-preserving path", got)
	}
}
