package acceptance_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/gates"
	"github.com/samuelnp/centinela/internal/verdict"
	"github.com/samuelnp/centinela/internal/verify"
	"github.com/samuelnp/centinela/internal/workflow"
)

// Acceptance: specs/headless-governance.feature

// Scenario: Verdict run info snapshots workflow and config provenance
func TestHG_RunInfoProvenance(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "1")
	pkt := hgAssemble([]gates.Result{hgPassGate()}, hgVerify(), nil)
	r := pkt.Run
	if r.Feature != "headless-governance" || r.Step != "validate" || r.Profile != "strict" {
		t.Fatalf("run = %+v", r)
	}
	if r.Archetype != "canonical" || r.DriverModel != "claude-opus" || !r.Headless || r.GeneratedAt != hgNow {
		t.Fatalf("provenance wrong: %+v", r)
	}
}

// Scenario: Verdict JSON is deterministic for fixed inputs and injected timestamp
func TestHG_DeterministicJSON(t *testing.T) {
	t.Setenv("CENTINELA_HEADLESS", "")
	ev := []verdict.EvidLine{
		{Role: "qa-senior", Step: "tests", Status: "done", Path: ".workflow/headless-governance-qa-senior.json"},
		{Role: "big-thinker", Step: "plan", Status: "done", Path: ".workflow/headless-governance-big-thinker.json"},
	}
	v := hgVerify(verify.Check{Role: "qa-senior", Claim: "tests-pass", Status: verify.StatusPass})
	a, _ := json.MarshalIndent(hgAssemble([]gates.Result{hgPassGate()}, v, ev), "", "  ")
	b, _ := json.MarshalIndent(hgAssemble([]gates.Result{hgPassGate()}, v, ev), "", "  ")
	if string(a) != string(b) {
		t.Fatal("two runs must be byte-identical for fixed inputs + injected Now")
	}
}

// Scenario: Verdict evidence index lists every on-disk role evidence for the feature
func TestHG_EvidenceIndexLists(t *testing.T) {
	d := t.TempDir()
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) })      //nolint:errcheck
	os.Chdir(d)                              //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0o755) //nolint:errcheck
	write := func(role, body string) {
		os.WriteFile(filepath.Join(workflow.WorkflowDir, "feat-"+role+".json"), []byte(body), 0o644) //nolint:errcheck
	}
	write("qa-senior", `{"feature":"feat","step":"tests","role":"qa-senior","status":"done","generatedAt":"2026-06-12T00:00:00Z","handoffTo":"validation-specialist"}`)
	write("big-thinker", `{"feature":"feat","step":"plan","role":"big-thinker","status":"done","generatedAt":"2026-06-11T00:00:00Z","handoffTo":"feature-specialist"}`)
	idx := verdict.EvidenceIndex("feat")
	if len(idx) != 2 || idx[0].Role != "big-thinker" || idx[1].Role != "qa-senior" {
		t.Fatalf("index not sorted/complete: %+v", idx)
	}
	if idx[0].HandoffTo != "feature-specialist" || idx[0].Path != ".workflow/feat-big-thinker.json" {
		t.Fatalf("entry fields wrong: %+v", idx[0])
	}
}

// Scenario: Verdict on a feature with no on-disk evidence yields an empty evidence index
func TestHG_EvidenceIndexEmpty(t *testing.T) {
	d := t.TempDir()
	old, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(old) })      //nolint:errcheck
	os.Chdir(d)                              //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0o755) //nolint:errcheck
	idx := verdict.EvidenceIndex("feat")
	if idx == nil || len(idx) != 0 {
		t.Fatalf("want empty non-nil slice, got %v", idx)
	}
	pkt := verdict.AssembleVerdict("feat", &config.Config{}, hgWf(), hgDeps([]gates.Result{hgPassGate()}, hgVerify(), idx))
	if _, err := json.Marshal(pkt); err != nil {
		t.Fatalf("packet must still marshal: %v", err)
	}
}
