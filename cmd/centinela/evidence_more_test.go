package main

import (
	"os"
	"strings"
	"testing"
)

func TestEvidenceInitListsActiveOnUnknown(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	writeFakeWorkflow(t, "beta")
	// "planner" (non-retired) so this exercises requireKnownFeature, not D7's
	// retired-role refusal, which now runs first and would otherwise mask it.
	err := runEvidenceInit(nil, []string{"ghost", "planner"})
	if err == nil {
		t.Fatal("expected error")
	}
	got := err.Error()
	if !strings.Contains(got, "alpha") || !strings.Contains(got, "beta") {
		t.Fatalf("active features not listed: %v", got)
	}
}

func TestEvidenceReadUnknownRoleRejected(t *testing.T) {
	chdirEvidenceTemp(t)
	if err := runEvidenceRead(nil, []string{"alpha", "ghost"}); err == nil {
		t.Fatal("expected unknown role error")
	}
}

func TestEvidenceAppendUnknownRoleRejected(t *testing.T) {
	chdirEvidenceTemp(t)
	if err := runEvidenceAppend(nil, []string{"alpha", "ghost", "outputs", "x"}); err == nil {
		t.Fatal("expected unknown role error")
	}
}

func TestEvidenceRepairUnknownFeatureNoop(t *testing.T) {
	chdirEvidenceTemp(t)
	out := captureStdout(t, func() {
		if err := runEvidenceRepair(nil, []string{"ghost"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "no orphaned") {
		t.Fatalf("unexpected: %q", out)
	}
}

func TestEvidenceSchemaReturnsValidJSON(t *testing.T) {
	out := captureStdout(t, func() {
		if err := runEvidenceSchema(nil, []string{"feature-specialist"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, `"step": "plan"`) {
		t.Fatalf("expected step plan in schema, got: %q", out)
	}
}

func TestEvidenceSetMissingEvidenceCreatesNothing(t *testing.T) {
	chdirEvidenceTemp(t)
	err := runEvidenceSet(nil, []string{"alpha", "big-thinker", "status", "done"})
	if err == nil || !strings.Contains(err.Error(), "evidence not found") {
		t.Fatalf("expected missing-evidence, got %v", err)
	}
	if _, err := os.Stat(".workflow/alpha-big-thinker.json"); !os.IsNotExist(err) {
		t.Fatal("file should not be created")
	}
}
