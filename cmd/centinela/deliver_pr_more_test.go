package main

import (
	"errors"
	"strings"
	"testing"
)

// cleanPushStub makes status report clean and push succeed, so runDeliverPR
// reaches the gh stage. addOrigin must be true for HasOriginRemote to pass.
func cleanPushStub(t *testing.T) {
	t.Helper()
	stubGitDeliver(t, func(args ...string) (string, error) {
		return "", nil // clean status, successful push
	})
}

func stubGH(t *testing.T, available bool, url string, err error) {
	t.Helper()
	pa, pc := ghAvailable, ghCreatePR
	ghAvailable = func() bool { return available }
	ghCreatePR = func(string, string, string) (string, error) { return url, err }
	t.Cleanup(func() { ghAvailable, ghCreatePR = pa, pc })
}

// TestRunDeliverPROpensPR: clean tree + origin + gh available → PR opened, exit 0.
func TestRunDeliverPROpensPR(t *testing.T) {
	deliverRepo(t, true)
	cleanPushStub(t)
	stubGH(t, true, "https://github.com/o/r/pull/7", nil)
	out := captureStdout(t, func() {
		if err := runDeliverPR(nil, "feat"); err != nil {
			t.Fatalf("expected success, got %v", err)
		}
	})
	if !strings.Contains(out, "Opened pull request") || !strings.Contains(out, "pull/7") {
		t.Fatalf("expected PR URL reported:\n%s", out)
	}
}

// TestRunDeliverPRGhAbsent: gh missing → branch pushed, manual instructions, non-zero.
func TestRunDeliverPRGhAbsent(t *testing.T) {
	deliverRepo(t, true)
	cleanPushStub(t)
	stubGH(t, false, "", nil)
	var err error
	out := captureStdout(t, func() { err = runDeliverPR(nil, "feat") })
	if err == nil || !strings.Contains(err.Error(), "gh CLI unavailable") {
		t.Fatalf("gh absent should exit non-zero, got %v", err)
	}
	if !strings.Contains(out, "Pushed") || !strings.Contains(out, "manually") {
		t.Fatalf("expected push + manual instructions:\n%s", out)
	}
	if strings.Contains(out, "Opened pull request") {
		t.Fatalf("must not claim a PR was opened:\n%s", out)
	}
}

// TestRunDeliverPRGhFails: gh present but `gh pr create` fails → surfaced error.
func TestRunDeliverPRGhFails(t *testing.T) {
	deliverRepo(t, true)
	cleanPushStub(t)
	stubGH(t, true, "boom", errors.New("gh exited 1"))
	if err := runDeliverPR(nil, "feat"); err == nil || !strings.Contains(err.Error(), "gh pr create failed") {
		t.Fatalf("gh failure should surface, got %v", err)
	}
}
