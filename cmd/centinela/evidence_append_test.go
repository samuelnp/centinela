package main

import (
	"os"
	"strings"
	"testing"
)

func TestEvidenceAppendDedups(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	if err := runEvidenceInit(nil, []string{"alpha", "big-thinker"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := runEvidenceAppend(nil, []string{"alpha", "big-thinker", "outputs", "docs/plans/alpha.md"}); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	data, _ := os.ReadFile(".workflow/alpha-big-thinker.json")
	if got := strings.Count(string(data), "docs/plans/alpha.md"); got != 1 {
		t.Fatalf("expected dedup to 1, got %d in %s", got, data)
	}
}

func TestEvidenceAppendRejectsScalarField(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	if err := runEvidenceInit(nil, []string{"alpha", "big-thinker"}); err != nil {
		t.Fatal(err)
	}
	err := runEvidenceAppend(nil, []string{"alpha", "big-thinker", "status", "done"})
	if err == nil || !strings.Contains(err.Error(), "not appendable") {
		t.Fatalf("expected non-appendable error, got %v", err)
	}
}

func TestEvidenceAppendRequiresInit(t *testing.T) {
	chdirEvidenceTemp(t)
	err := runEvidenceAppend(nil, []string{"alpha", "big-thinker", "outputs", "x"})
	if err == nil || !strings.Contains(err.Error(), "evidence not found") {
		t.Fatalf("expected missing-evidence error, got %v", err)
	}
}
