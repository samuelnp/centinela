package worktree

import (
	"errors"
	"strings"
	"testing"
)

// A bare primary tree is a hard refusal — merging in it is impossible.
func TestPrimaryTree_BarePrimary_Refused(t *testing.T) {
	porcelainStub(t, "worktree /repo.git\nbare\n\nworktree /repo/.worktrees/feat\nHEAD def\n\n")
	_, err := PrimaryTree(".")
	if err == nil || !strings.Contains(err.Error(), "bare") {
		t.Fatalf("bare primary must be refused, got: %v", err)
	}
}

// Empty output means git answered but nothing is parseable: never guess.
func TestPrimaryTree_EmptyOutput_Refused(t *testing.T) {
	porcelainStub(t, "")
	_, err := PrimaryTree(".")
	if err == nil || !strings.Contains(err.Error(), "no worktree entry") {
		t.Fatalf("empty porcelain must be refused, got: %v", err)
	}
}

// Garbage output with no `worktree ` line is equally unparseable.
func TestPrimaryTree_GarbageOutput_Refused(t *testing.T) {
	porcelainStub(t, "something unexpected\nanother line\n")
	_, err := PrimaryTree(".")
	if err == nil || !strings.Contains(err.Error(), "no worktree entry") {
		t.Fatalf("garbage porcelain must be refused, got: %v", err)
	}
}

// Runner failures (e.g. not a git repository) propagate with context.
func TestPrimaryTree_RunnerError_Propagated(t *testing.T) {
	stubGit(t, func(string, ...string) ([]byte, error) {
		return []byte("fatal: not a git repository"), errors.New("exit status 128")
	})
	_, err := PrimaryTree(".")
	if err == nil || !strings.Contains(err.Error(), "cannot resolve primary working tree") {
		t.Fatalf("runner error must surface the refusal prefix, got: %v", err)
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("runner error must carry git's message, got: %v", err)
	}
}
