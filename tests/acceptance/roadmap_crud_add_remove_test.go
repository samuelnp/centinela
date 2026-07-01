package acceptance_test

// Acceptance: specs/roadmap-crud-add-remove.feature
// Scenario: add creates a draft in a chosen schedulable phase and validate stays PASS
// Scenario: a freshly-added draft simultaneously satisfies all four draft readers
// Scenario Outline: add rejects invalid input and leaves roadmap.json byte-identical
// Scenario: add against an empty roadmap errors "unknown phase"
//
// Note: the full `centinela start <draft>` refusal is asserted at the guard
// level in cmd/centinela/start_guard_draft_test.go (driving `start` needs a
// PROJECT.md + centinela.toml); here the draft's exemption is asserted through
// the roadmap --json / ready / validate readers.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// crudSeeded writes a temp project whose roadmap has a scored, non-draft "seed"
// feature plus matching analysis/quality artifacts, so validate PASSes.
func crudSeeded(t *testing.T) string {
	t.Helper()
	body := `{"phases":[{"name":"Phase 1: Foundations","features":[{"name":"seed"}]}]}`
	d := rmcProject(t, body)
	wf := filepath.Join(d, ".workflow")
	write := func(n, c string) {
		if err := os.WriteFile(filepath.Join(wf, n), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("roadmap-analysis.json", `{"role":"senior-product-manager","features":[{"name":"seed"}]}`)
	write("roadmap-quality.json", `{"role":"roadmap-quality-evaluator","threshold":9,"features":[`+
		`{"name":"seed","scores":{"acceptanceCriteria":9,"userValue":9,"definitionClarity":9,`+
		`"dependencies":9,"effortEstimation":9,"overall":9},"summary":"s"}]}`)
	write("roadmap-analysis.md", "# a\n")
	write("roadmap-quality.md", "# q\n")
	return d
}

// TestAcc_AddDraftValidatesAndReaders drives add, then validate + the four readers.
func TestAcc_AddDraftValidatesAndReaders(t *testing.T) {
	d := crudSeeded(t)
	if _, _, code := rmcRun(t, d, "roadmap", "add", "new-widget", "--phase", "Phase 1: Foundations"); code != 0 {
		t.Fatalf("add exit=%d", code)
	}
	if _, _, code := rmcRun(t, d, "roadmap", "validate"); code != 0 {
		t.Fatalf("validate must stay PASS after add, exit=%d", code)
	}
	out, _, code := rmcRun(t, d, "roadmap", "--json")
	if code != 0 {
		t.Fatalf("--json exit=%d", code)
	}
	var v roadmap.RoadmapView
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	var found roadmap.FeatureView
	for _, f := range v.Phases[0].Features {
		if f.Name == "new-widget" {
			found = f
		}
	}
	if !found.Draft || found.Readiness != "draft" {
		t.Fatalf("draft reader: %+v", found)
	}
	if v.Counts.Planned != 1 { // only seed, draft excluded
		t.Fatalf("draft must be excluded from counts: %+v", v.Counts)
	}
	ready, _, _ := rmcRun(t, d, "roadmap", "ready", "--json")
	if strings.Contains(ready, "new-widget") {
		t.Fatalf("ready must exclude the draft: %s", ready)
	}
}

// TestAcc_AddRejectionsByteIdentical drives the reject rows through the binary.
func TestAcc_AddRejectionsByteIdentical(t *testing.T) {
	rows := []struct{ args []string }{
		{[]string{"roadmap", "add", "Not_Kebab!", "--phase", "Phase 1: Foundations"}},
		{[]string{"roadmap", "add", "seed", "--phase", "Phase 1: Foundations"}},
		{[]string{"roadmap", "add", "x", "--phase", "Backlog"}},
		{[]string{"roadmap", "add", "x", "--phase", "Phase 1: Foundations", "--depends-on", "x"}},
	}
	for _, r := range rows {
		d := crudSeeded(t)
		p := filepath.Join(d, ".workflow", "roadmap.json")
		before, _ := os.ReadFile(p)
		if _, _, code := rmcRun(t, d, r.args...); code == 0 {
			t.Fatalf("row %v must be rejected", r.args)
		}
		after, _ := os.ReadFile(p)
		if string(before) != string(after) {
			t.Fatalf("row %v must leave roadmap.json byte-identical", r.args)
		}
	}
}
