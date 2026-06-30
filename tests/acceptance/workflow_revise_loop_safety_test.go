// Acceptance: specs/workflow-revise-loop.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: Safety — invalidation never touches source, test, or docs files
func TestRL_SafetyNoSourceDeletion(t *testing.T) {
	dir := rlDir(t)
	rlState(t, dir, "my-feature", rlCanonical, "validate")
	keep := []string{
		"internal/myfeature/service.go",
		"tests/unit/myfeature/service_test.go",
		"docs/features/workflow-revise-loop.md",
		"docs/plans/workflow-revise-loop.md",
	}
	for _, p := range keep {
		rlWrite(t, dir, p, "content")
	}
	rlWrite(t, dir, ".workflow/my-feature-validation-specialist.json", "x")
	out, code := runCent(t, buildCent(t), dir,
		"revise", "my-feature", "--to", "code", "--reason", "safety check")
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, out)
	}
	mustExist(t, dir, keep...)
	mustGone(t, dir, "my-feature-validation-specialist.json")
}

// Scenario: Idempotency — invalidating already-absent evidence is not an error
func TestRL_IdempotentInvalidation(t *testing.T) {
	dir := rlDir(t)
	rlState(t, dir, "my-feature", rlCanonical, "validate")
	// No gatekeeper evidence is seeded — invalidation must not complain.
	out, code := runCent(t, buildCent(t), dir,
		"revise", "my-feature", "--to", "code", "--reason", "idempotent")
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, out)
	}
	if strings.Contains(strings.ToLower(out), "no such file") ||
		strings.Contains(strings.ToLower(out), "not found") {
		t.Fatalf("absent evidence must not surface an error: %s", out)
	}
}

// Scenario: Multiple rewinds accumulate in the revision log
func TestRL_MultipleRewindsAccumulate(t *testing.T) {
	dir := rlDir(t)
	rlWrite(t, dir, ".workflow/my-feature.json", `{"feature":"my-feature",`+
		`"currentStep":"validate","stepOrder":["plan","code","tests","validate","docs"],`+
		`"steps":{"plan":{"status":"done"},"code":{"status":"done"},`+
		`"tests":{"status":"done"},"validate":{"status":"in-progress"},"docs":{"status":"pending"}},`+
		`"startedAt":"2026-06-30T00:00:00Z","revisions":[`+
		`{"from":"validate","to":"code","reason":"first rewind","at":"2026-06-30T00:00:00Z"}]}`)
	out, code := runCent(t, buildCent(t), dir,
		"revise", "my-feature", "--to", "code", "--reason", "second rewind")
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, out)
	}
	revs, _ := rlLoad(t, dir, "my-feature")["revisions"].([]any)
	if len(revs) != 2 {
		t.Fatalf("revisions = %d, want 2", len(revs))
	}
	first, _ := revs[0].(map[string]any)
	second, _ := revs[1].(map[string]any)
	if first["reason"] != "first rewind" || second["reason"] != "second rewind" {
		t.Fatalf("revision log = %v", revs)
	}
}

// Scenario: Internal feature code-step invalidation excludes ux-ui-specialist
func TestRL_InternalFeatureNoUXInvalidation(t *testing.T) {
	dir := rlDir(t)
	// Internal feature (no docs/features/<f>.md) sitting on tests.
	rlState(t, dir, "my-feature", rlCanonical, "tests")
	rlWrite(t, dir, ".workflow/my-feature-qa-senior.json", "x")
	out, code := runCent(t, buildCent(t), dir,
		"revise", "my-feature", "--to", "code", "--reason", "internal fix")
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, out)
	}
	if rlLoad(t, dir, "my-feature")["currentStep"] != "code" {
		t.Fatal("current step must be code")
	}
	if strings.Contains(out, "ux-ui-specialist") {
		t.Fatalf("internal code-step invalidation must not reference ux-ui: %s", out)
	}
	// The re-opened tests step's qa-senior evidence is shed.
	mustGone(t, dir, "my-feature-qa-senior.json")
}
