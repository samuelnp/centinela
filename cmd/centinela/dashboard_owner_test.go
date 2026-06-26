package main

import (
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

// TestGitOwner_DefaultUnknownOnBogusBranch exercises the real seam against a
// branch that does not exist in any repo, which makes git log fail → "unknown".
func TestGitOwner_DefaultUnknownOnBogusBranch(t *testing.T) {
	dir := t.TempDir() // not a git repo
	if got := gitOwner(dir, "no-such-branch-xyz"); got != "unknown" {
		t.Fatalf("bogus branch => %q, want unknown", got)
	}
}

func TestDashboardOwners_MapsActiveAndSkipsNil(t *testing.T) {
	prev := gitOwner
	t.Cleanup(func() { gitOwner = prev })
	gitOwner = func(_, feature string) string {
		if feature == "beta" {
			return "Carol"
		}
		return "unknown"
	}
	active := []*workflow.Workflow{
		{Feature: "alpha"},
		nil,
		{Feature: "beta"},
	}
	owners := dashboardOwners(active)
	if len(owners) != 2 {
		t.Fatalf("nil workflow must be skipped: %+v", owners)
	}
	if owners["alpha"] != "unknown" {
		t.Fatalf("alpha owner: %q", owners["alpha"])
	}
	if owners["beta"] != "Carol" {
		t.Fatalf("beta owner: %q", owners["beta"])
	}
}
