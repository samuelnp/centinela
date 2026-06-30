package main

import (
	"errors"
	"strings"
	"testing"
)

// TestCommitChangelogAddFails covers the `git add` failure branch.
func TestCommitChangelogAddFails(t *testing.T) {
	deliverRepo(t, true)
	seedChangelogSources(t, "feat")
	stubGitDeliver(t, func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "add" {
			return "denied", errors.New("add boom")
		}
		return "", nil
	})
	if err := commitChangelog("feat"); err == nil ||
		!strings.Contains(err.Error(), "git add") {
		t.Fatalf("add failure should surface, got %v", err)
	}
}

// TestCommitChangelogCommitFails covers the `git commit` failure branch.
func TestCommitChangelogCommitFails(t *testing.T) {
	deliverRepo(t, true)
	seedChangelogSources(t, "feat")
	stubGitDeliver(t, func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "commit" {
			return "denied", errors.New("commit boom")
		}
		return "", nil
	})
	if err := commitChangelog("feat"); err == nil ||
		!strings.Contains(err.Error(), "git commit changelog") {
		t.Fatalf("commit failure should surface, got %v", err)
	}
}
