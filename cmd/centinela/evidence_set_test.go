package main

import (
	"os"
	"strings"
	"testing"
)

func TestEvidenceSetRequiresInit(t *testing.T) {
	chdirEvidenceTemp(t)
	err := runEvidenceSet(nil, []string{"alpha", "big-thinker", "status", "done"})
	if err == nil || !strings.Contains(err.Error(), "evidence not found") {
		t.Fatalf("expected missing-evidence, got %v", err)
	}
}

func TestEvidenceSetMutatesField(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	if err := runEvidenceInit(nil, []string{"alpha", "big-thinker"}); err != nil {
		t.Fatal(err)
	}
	if err := runEvidenceSet(nil, []string{"alpha", "big-thinker", "handoffTo", "feature-specialist"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(".workflow/alpha-big-thinker.json")
	if !strings.Contains(string(data), `"handoffTo": "feature-specialist"`) {
		t.Fatalf("set did not persist: %s", data)
	}
	matches, _ := os.ReadDir(".workflow")
	for _, m := range matches {
		if strings.HasSuffix(m.Name(), ".tmp") {
			t.Fatalf("temp file remained: %s", m.Name())
		}
	}
}

func TestEvidenceSetExtraPath(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	if err := runEvidenceInit(nil, []string{"alpha", "big-thinker"}); err != nil {
		t.Fatal(err)
	}
	if err := runEvidenceSet(nil, []string{"alpha", "big-thinker", "extra.note", "reviewed by sam"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(".workflow/alpha-big-thinker.json")
	if !strings.Contains(string(data), `"note": "reviewed by sam"`) {
		t.Fatalf("extra not stored: %s", data)
	}
}

func TestEvidenceSetUnknownRole(t *testing.T) {
	chdirEvidenceTemp(t)
	err := runEvidenceSet(nil, []string{"alpha", "ghost", "status", "done"})
	if err == nil || !strings.Contains(err.Error(), "unknown role") {
		t.Fatalf("expected unknown role, got %v", err)
	}
}
