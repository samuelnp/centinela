package main

import (
	"strings"
	"testing"
)

func stubGitDeliver(t *testing.T, fn func(args ...string) (string, error)) {
	t.Helper()
	prev := gitDeliver
	gitDeliver = fn
	t.Cleanup(func() { gitDeliver = prev })
}

// TestRunDeliverPRNoOrigin refuses with no origin remote and never pushes.
func TestRunDeliverPRNoOrigin(t *testing.T) {
	deliverRepo(t, false)
	pushed := false
	stubGitDeliver(t, func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "push" {
			pushed = true
		}
		return "", nil
	})
	if err := runDeliverPR(nil, "feat"); err == nil || !strings.Contains(err.Error(), "no origin remote") {
		t.Fatalf("no origin should refuse, got %v", err)
	}
	if pushed {
		t.Fatal("must not push when origin is absent")
	}
}

// TestRunDeliverPRDirtyTree refuses to push when the worktree has uncommitted
// changes (origin present via real git).
func TestRunDeliverPRDirtyTree(t *testing.T) {
	deliverRepo(t, true)
	pushed := false
	stubGitDeliver(t, func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "status" {
			return " M file.go", nil
		}
		if len(args) > 0 && args[0] == "push" {
			pushed = true
		}
		return "", nil
	})
	if err := runDeliverPR(nil, "feat"); err == nil || !strings.Contains(err.Error(), "uncommitted changes") {
		t.Fatalf("dirty tree should be refused, got %v", err)
	}
	if pushed {
		t.Fatal("must not push with a dirty tree")
	}
}
