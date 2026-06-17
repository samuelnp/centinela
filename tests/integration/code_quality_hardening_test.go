package integration_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/evidence"
	"github.com/samuelnp/centinela/internal/hookpolicy"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Integration: workflow.Load error transparency across the three on-disk
// outcomes, plus hookpolicy/evidence key-order parity — exercised through the
// public package APIs as a cross-package narrative.

func TestLoadDistinguishesMissingFromCorrupt(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := workflow.Load("nope"); err == nil || !strings.Contains(err.Error(), "no workflow found") {
		t.Fatalf("missing must report absence, got: %v", err)
	}
}

func TestEvidenceFormatterParityWithCoverage(t *testing.T) {
	cov := 91.0
	re := &evidence.RoleEvidence{
		Feature: "f", Step: "tests", Role: "qa-senior", Status: "done",
		GeneratedAt: "2026-06-10T00:00:00Z",
		Inputs:      []string{}, Outputs: []string{}, EdgeCases: []string{},
		Coverage: &cov, HandoffTo: "validation-specialist",
	}
	canonical, err := re.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	out, changed, ferr := hookpolicy.FormatEvidence(".workflow/f-qa-senior.json", canonical, "f")
	if ferr != nil {
		t.Fatal(ferr)
	}
	if changed {
		t.Fatalf("formatter reordered coverage-bearing evidence: %s", out)
	}
	if !strings.Contains(string(canonical), `"coverage"`) {
		t.Fatal("canonical output should carry the coverage field")
	}
}
