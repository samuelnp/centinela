package main

import (
	"strings"
	"testing"
)

func TestEvidenceSetBadFieldSurfaces(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	if err := runEvidenceInit(nil, []string{"alpha", "big-thinker"}); err != nil {
		t.Fatal(err)
	}
	err := runEvidenceSet(nil, []string{"alpha", "big-thinker", "outputs", "x"})
	if err == nil || !strings.Contains(err.Error(), "list") {
		t.Fatalf("expected list-rejection from SetField, got %v", err)
	}
}

func TestEvidenceAppendBadFieldSurfaces(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	if err := runEvidenceInit(nil, []string{"alpha", "big-thinker"}); err != nil {
		t.Fatal(err)
	}
	err := runEvidenceAppend(nil, []string{"alpha", "big-thinker", "feature", "x"})
	if err == nil || !strings.Contains(err.Error(), "not appendable") {
		t.Fatal("expected not-appendable error")
	}
}

func TestEvidenceReadBadFieldSurfaces(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	if err := runEvidenceInit(nil, []string{"alpha", "big-thinker"}); err != nil {
		t.Fatal(err)
	}
	evidenceReadField = "bogus"
	t.Cleanup(func() { evidenceReadField = "" })
	err := runEvidenceRead(nil, []string{"alpha", "big-thinker"})
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown-field, got %v", err)
	}
}
