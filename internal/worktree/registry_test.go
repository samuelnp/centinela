package worktree

import (
	"errors"
	"strings"
	"testing"
)

const twoTreePorcelain = "worktree /repo\nHEAD aaa\nbranch refs/heads/main\n\n" +
	"worktree /elsewhere/high score\nHEAD bbb\nbranch refs/heads/high-score\n\n"

// The registry — not the .worktrees/<feature> convention — decides whether a
// worktree is still live. Paths may contain spaces.
func TestFindRegisteredWorktree_MatchesBranchOutsideConvention(t *testing.T) {
	path, ok := findRegisteredWorktree(twoTreePorcelain, "high-score")
	if !ok || path != "/elsewhere/high score" {
		t.Fatalf("findRegisteredWorktree = %q, %v", path, ok)
	}
}

func TestFindRegisteredWorktree_UnknownBranchNotFound(t *testing.T) {
	if path, ok := findRegisteredWorktree(twoTreePorcelain, "ghost"); ok {
		t.Fatalf("unknown branch must not match, got %q", path)
	}
}

// A prunable block is a stale administrative record, not a worktree on disk.
func TestFindRegisteredWorktree_PrunableEntryIsNotLive(t *testing.T) {
	out := "worktree /repo\nbranch refs/heads/main\n\n" +
		"worktree /gone\nbranch refs/heads/high-score\nprunable gitdir file points to non-existent location\n\n"
	if path, ok := findRegisteredWorktree(out, "high-score"); ok {
		t.Fatalf("prunable entry must not count as live, got %q", path)
	}
}

func TestFindRegisteredWorktree_ToleratesCRLF(t *testing.T) {
	out := "worktree /repo\r\nbranch refs/heads/main\r\n\r\n" +
		"worktree /wt\r\nbranch refs/heads/feat\r\n\r\n"
	if path, ok := findRegisteredWorktree(out, "feat"); !ok || path != "/wt" {
		t.Fatalf("CRLF porcelain must parse, got %q %v", path, ok)
	}
}

// A registry lookup failure is an error, never a silent "nothing registered":
// that is exactly how a false removal claim would slip through.
func TestRegisteredWorktree_RunnerErrorPropagates(t *testing.T) {
	stubGit(t, func(string, ...string) ([]byte, error) {
		return []byte("not a git repository"), errors.New("exit status 128")
	})
	_, ok, err := registeredWorktree("/repo", "feat")
	if err == nil || ok {
		t.Fatalf("runner failure must surface, got ok=%v err=%v", ok, err)
	}
	if !strings.Contains(err.Error(), "git worktree list failed") {
		t.Fatalf("error must name the failing command: %v", err)
	}
}
