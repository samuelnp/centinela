// Acceptance: specs/workflow-revise-loop.feature
package acceptance_test

import (
	"strings"
	"testing"
)

var rlCanonical = []string{"plan", "code", "tests", "validate", "docs"}

// Scenario: Happy path — validate step is revised back to code
func TestRL_HappyPathValidateToCode(t *testing.T) {
	dir := rlDir(t)
	rlState(t, dir, "my-feature", rlCanonical, "validate")
	for _, r := range []string{"senior-engineer", "gatekeeper", "validation-specialist", "qa-senior"} {
		rlWrite(t, dir, ".workflow/my-feature-"+r+".json", "x")
		rlWrite(t, dir, ".workflow/my-feature-"+r+".md", "x")
	}
	rlWrite(t, dir, ".workflow/my-feature-edge-cases.md", "x")
	rlWrite(t, dir, "internal/myfeature/handler.go", "package myfeature")

	out, code := runCent(t, buildCent(t), dir,
		"revise", "my-feature", "--to", "code", "--reason", "bug found in handler")
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, out)
	}
	m := rlLoad(t, dir, "my-feature")
	if m["currentStep"] != "code" {
		t.Fatalf("current = %v", m["currentStep"])
	}
	if rlStep(m, "code")["status"] != "in-progress" {
		t.Fatalf("code = %v", rlStep(m, "code"))
	}
	for _, s := range []string{"tests", "validate", "docs"} {
		if rlStep(m, s)["status"] != "pending" {
			t.Fatalf("%s = %v", s, rlStep(m, s))
		}
	}
	mustGone(t, dir, "my-feature-gatekeeper.json", "my-feature-validation-specialist.md",
		"my-feature-edge-cases.md")
	mustExist(t, dir, ".workflow/my-feature-senior-engineer.json", "internal/myfeature/handler.go")
	revs, _ := m["revisions"].([]any)
	if len(revs) != 1 {
		t.Fatalf("revisions = %v", revs)
	}
	r0, _ := revs[0].(map[string]any)
	if r0["from"] != "validate" || r0["to"] != "code" || r0["reason"] != "bug found in handler" {
		t.Fatalf("revision = %v", r0)
	}
}

// Scenario: Archetype-awareness — rewind respects hotfix step order
func TestRL_ArchetypeHotfixOrder(t *testing.T) {
	dir := rlDir(t)
	rlState(t, dir, "hotfix-one", []string{"code", "tests", "validate"}, "validate")
	out, code := runCent(t, buildCent(t), dir,
		"revise", "hotfix-one", "--to", "code", "--reason", "hotfix regression")
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, out)
	}
	m := rlLoad(t, dir, "hotfix-one")
	if m["currentStep"] != "code" {
		t.Fatalf("current = %v", m["currentStep"])
	}
	for _, s := range []string{"tests", "validate"} {
		if rlStep(m, s)["status"] != "pending" {
			t.Fatalf("%s = %v", s, rlStep(m, s))
		}
	}
	steps, _ := m["steps"].(map[string]any)
	if _, ok := steps["plan"]; ok {
		t.Fatal("hotfix must have no plan step")
	}
	if _, ok := steps["docs"]; ok {
		t.Fatal("hotfix must have no docs step")
	}
	if !strings.Contains(out, "code") {
		t.Fatalf("output should name the new current step: %s", out)
	}
}
