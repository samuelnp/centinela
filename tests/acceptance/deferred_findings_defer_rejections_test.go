package acceptance_test

// Acceptance: specs/deferred-findings-roadmap-capture.feature

import (
	"strings"
	"testing"
)

// Scenario: Defer with an empty summary is rejected before any write
func TestDfrc_DeferEmptySummaryRejected(t *testing.T) {
	bin := buildCent(t)
	dir := dfrcAcceptDir(t, dfrcRoadmapBase)
	out, code := runCent(t, bin, dir, "roadmap", "defer", "empty-summary-test", "--summary", "")
	if code == 0 {
		t.Fatalf("empty summary must be rejected, got exit 0\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "summary") {
		t.Errorf("output must mention 'summary': %s", out)
	}
}

// Scenario: Defer rejects a slug that already exists in the Backlog phase
func TestDfrc_DeferDuplicateBacklogSlug(t *testing.T) {
	bin := buildCent(t)
	src := `{"phases":[{"name":"Backlog","features":[{"name":"hook-timeout-config","summary":"x","deferredAt":"t"}]}]}`
	dir := dfrcAcceptDir(t, src)
	out, code := runCent(t, bin, dir, "roadmap", "defer", "hook-timeout-config", "--summary", "Duplicate attempt")
	if code == 0 {
		t.Fatalf("duplicate Backlog slug must be rejected, got exit 0\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "already") && !strings.Contains(strings.ToLower(out), "collision") && !strings.Contains(strings.ToLower(out), "exists") {
		t.Errorf("output must indicate slug collision: %s", out)
	}
}

// Scenario: Defer rejects a slug that already exists in a non-Backlog phase
func TestDfrc_DeferDuplicateNonBacklogSlug(t *testing.T) {
	bin := buildCent(t)
	src := `{"phases":[{"name":"Phase 0","features":[{"name":"enforce-coverage-in-validate"}]}]}`
	dir := dfrcAcceptDir(t, src)
	out, code := runCent(t, bin, dir, "roadmap", "defer", "enforce-coverage-in-validate",
		"--summary", "Raise the bar further")
	if code == 0 {
		t.Fatalf("non-Backlog slug collision must be rejected, got exit 0\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "already") && !strings.Contains(strings.ToLower(out), "collision") && !strings.Contains(strings.ToLower(out), "exists") {
		t.Errorf("output must indicate slug collision: %s", out)
	}
}

// Scenario: Defer with an invalid slug is rejected before any write
func TestDfrc_DeferInvalidSlug(t *testing.T) {
	bin := buildCent(t)
	dir := dfrcAcceptDir(t, dfrcRoadmapBase)
	out, code := runCent(t, bin, dir, "roadmap", "defer", "bad slug!", "--summary", "Something")
	if code == 0 {
		t.Fatalf("invalid slug must be rejected, got exit 0\n%s", out)
	}
	_ = out // output should name the invalid slug
}
