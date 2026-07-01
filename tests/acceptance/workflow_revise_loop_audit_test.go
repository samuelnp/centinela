// Acceptance: specs/workflow-revise-loop.feature
package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: Audit — revision count and reason are visible in status
func TestRL_AuditVisibleInStatus(t *testing.T) {
	dir := rlDir(t)
	rlState(t, dir, "my-feature", rlCanonical, "code")
	// Bake two revisions directly into the persisted state.
	rlWrite(t, dir, ".workflow/my-feature.json", `{"feature":"my-feature",`+
		`"currentStep":"code","stepOrder":["plan","code","tests","validate","docs"],`+
		`"steps":{"plan":{"status":"done"},"code":{"status":"in-progress"},`+
		`"tests":{"status":"pending"},"validate":{"status":"pending"},"docs":{"status":"pending"}},`+
		`"startedAt":"2026-06-30T00:00:00Z","revisions":[`+
		`{"from":"validate","to":"code","reason":"first rewind","at":"2026-06-30T00:00:00Z"},`+
		`{"from":"validate","to":"code","reason":"second rewind","at":"2026-06-30T01:00:00Z"}]}`)
	out, code := runCent(t, buildCent(t), dir, "status", "my-feature")
	if code != 0 {
		t.Fatalf("status exit %d: %s", code, out)
	}
	if !strings.Contains(out, "Revisions") || !strings.Contains(out, "2") {
		t.Fatalf("status must show the revision count: %s", out)
	}
	if !strings.Contains(out, "second rewind") {
		t.Fatalf("status must show the latest reason inline: %s", out)
	}
}

// Scenario: Re-opened step CompletedAt is cleared
func TestRL_CompletedAtClearedOnReopen(t *testing.T) {
	dir := rlDir(t)
	// Canonical at validate: plan/code/tests carry a non-null completedAt.
	rlState(t, dir, "my-feature", rlCanonical, "validate")
	if rlStep(rlLoad(t, dir, "my-feature"), "tests")["completedAt"] == nil {
		t.Fatal("precondition: tests must start with a non-null completedAt")
	}
	out, code := runCent(t, buildCent(t), dir,
		"revise", "my-feature", "--to", "code", "--reason", "clear timestamps")
	if code != 0 {
		t.Fatalf("want exit 0, got %d: %s", code, out)
	}
	m := rlLoad(t, dir, "my-feature")
	for _, s := range []string{"tests", "validate", "code"} {
		if rlStep(m, s)["completedAt"] != nil {
			t.Fatalf("%s completedAt must be cleared: %v", s, rlStep(m, s))
		}
	}
}
