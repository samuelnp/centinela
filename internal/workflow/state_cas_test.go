package workflow

import (
	"strings"
	"testing"
)

// TestStaleSaveIsRefused reproduces the observed bug deterministically. wfA
// stands in for `complete` (loads, runs the whole validate gate, then saves)
// and wfB for a `route set` landing in between; wfA's minutes-old copy must be
// refused instead of silently discarding wfB's update.
func TestStaleSaveIsRefused(t *testing.T) {
	stateRepo(t)
	if err := Save(New("epsilon")); err != nil {
		t.Fatal(err)
	}
	wfA, err := Load("epsilon")
	if err != nil {
		t.Fatal(err)
	}
	wfB, err := Load("epsilon")
	if err != nil {
		t.Fatal(err)
	}

	wfB.SetModelRoute("senior-engineer", ModelRoute{Tier: "balanced"})
	if err := Save(wfB); err != nil {
		t.Fatalf("the first writer must win: %v", err)
	}

	wfA.CurrentStep = "docs"
	err = Save(wfA)
	if err == nil {
		t.Fatal("a stale save must be refused, not silently applied")
	}
	for _, want := range []string{FilePath("epsilon"), "changed on disk", "Re-run this command"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("conflict error must contain %q, got: %v", want, err)
		}
	}

	current, err := Load("epsilon")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := current.ModelRoutes["senior-engineer"]; !ok {
		t.Fatal("the refused save destroyed the other writer's model route")
	}
	if current.CurrentStep == "docs" {
		t.Fatal("the refused save was applied anyway")
	}
}

// The documented recovery: re-run, which re-loads and then succeeds, leaving
// both writers' changes present.
func TestReRunAfterConflictSucceeds(t *testing.T) {
	stateRepo(t)
	if err := Save(New("epsilon")); err != nil {
		t.Fatal(err)
	}
	stale, _ := Load("epsilon")
	fresh, _ := Load("epsilon")
	fresh.SetModelRoute("qa-senior", ModelRoute{Tier: "fast"})
	if err := Save(fresh); err != nil {
		t.Fatal(err)
	}
	stale.CurrentStep = "docs"
	if err := Save(stale); err == nil {
		t.Fatal("precondition: the stale save must be refused")
	}

	retry, err := Load("epsilon")
	if err != nil {
		t.Fatal(err)
	}
	retry.CurrentStep = "docs"
	if err := Save(retry); err != nil {
		t.Fatalf("re-running after a conflict must succeed: %v", err)
	}
	got, _ := Load("epsilon")
	if _, ok := got.ModelRoutes["qa-senior"]; !ok || got.CurrentStep != "docs" {
		t.Fatalf("both changes must survive the retry: %+v", got)
	}
}

// start and the autostart hook build their workflow with New/NewWithOrder and
// never load, so they carry no digest and must never hit a conflict check.
func TestNeverLoadedWorkflowSavesWithoutCheck(t *testing.T) {
	stateRepo(t)
	if err := Save(New("zeta")); err != nil {
		t.Fatal(err)
	}
	if err := Save(New("zeta")); err != nil {
		t.Fatalf("an unloaded workflow must save over an existing file: %v", err)
	}
	wf, err := Load("zeta")
	if err != nil || wf.Feature != "zeta" || wf.CurrentStep != "plan" {
		t.Fatalf("unexpected state: %+v (%v)", wf, err)
	}
}
