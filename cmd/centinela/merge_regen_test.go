package main

import (
	"errors"
	"strings"
	"testing"
)

// Scenario 8: a clean merge regenerates the documentation portal — the seam is
// invoked exactly once.
func TestRunMergeCleanInvokesPortalRegen(t *testing.T) {
	seedCleanMergeRepo(t, "omega")

	orig := docsPortalRegen
	defer func() { docsPortalRegen = orig }()
	called := 0
	docsPortalRegen = func() error { called++; return nil }

	if err := runMerge(nil, []string{"omega"}); err != nil {
		t.Fatalf("clean merge should succeed: %v", err)
	}
	if called != 1 {
		t.Fatalf("portal regen must run once on a clean merge, ran %d times", called)
	}
}

// Scenario 9: a portal-regen failure does not fail a clean merge — the merge
// still succeeds and a notice is reported.
func TestRunMergeCleanToleratesPortalRegenFailure(t *testing.T) {
	seedCleanMergeRepo(t, "omega")

	orig := docsPortalRegen
	defer func() { docsPortalRegen = orig }()
	docsPortalRegen = func() error { return errors.New("roadmap.json missing") }

	out := captureStdout(t, func() {
		if err := runMerge(nil, []string{"omega"}); err != nil {
			t.Fatalf("regen failure must not fail the merge: %v", err)
		}
	})
	if !strings.Contains(out, "notice: portal regen skipped") {
		t.Fatalf("a regen failure must report a notice, got: %s", out)
	}
}
