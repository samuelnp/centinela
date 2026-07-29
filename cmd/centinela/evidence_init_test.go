package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func chdirEvidenceTemp(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}

func writeFakeWorkflow(t *testing.T, feature string) {
	t.Helper()
	if err := workflow.Save(workflow.New(feature)); err != nil {
		t.Fatal(err)
	}
}

func TestEvidenceInitWritesSkeleton(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	if err := runEvidenceInit(nil, []string{"alpha", "planner"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(".workflow/alpha-planner.json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"feature": "alpha"`) {
		t.Fatalf("skeleton missing feature: %s", data)
	}
	if _, err := os.Stat(".workflow/alpha-planner.md"); err != nil {
		t.Fatalf("companion not written: %v", err)
	}
}

func TestEvidenceInitRejectsUnknownFeature(t *testing.T) {
	chdirEvidenceTemp(t)
	// "planner" (non-retired) so this exercises requireKnownFeature, not D7's
	// retired-role refusal, which now runs first and would otherwise mask it.
	err := runEvidenceInit(nil, []string{"ghost", "planner"})
	if err == nil || !strings.Contains(err.Error(), "unknown feature") {
		t.Fatalf("expected unknown-feature error, got %v", err)
	}
}

func TestEvidenceInitRejectsUnknownRole(t *testing.T) {
	chdirEvidenceTemp(t)
	writeFakeWorkflow(t, "alpha")
	err := runEvidenceInit(nil, []string{"alpha", "bogus-role"})
	if err == nil || !strings.Contains(err.Error(), "unknown role") {
		t.Fatalf("expected unknown-role error, got %v", err)
	}
}
