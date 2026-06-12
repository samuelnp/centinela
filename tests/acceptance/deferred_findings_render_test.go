package acceptance_test

// Acceptance: specs/deferred-findings-roadmap-capture.feature

import (
	"strings"
	"testing"
)

// Scenario: Backlog findings are shown in centinela roadmap output when present
func TestDfrc_BacklogShownInRoadmapOutput(t *testing.T) {
	bin := buildCent(t)
	src := `{"phases":[{"name":"Backlog","features":[
		{"name":"hook-timeout-config","summary":"Prewrite hook timeout","deferredAt":"t"},
		{"name":"doc-sync-reminder","summary":"Sync docs","deferredAt":"t"}]}]}`
	dir := dfrcAcceptDir(t, src)
	out, code := runCent(t, bin, dir, "roadmap")
	if code != 0 {
		t.Fatalf("roadmap exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "Backlog") {
		t.Error("output must contain Backlog section")
	}
	if !strings.Contains(out, "hook-timeout-config") {
		t.Error("output must mention hook-timeout-config")
	}
	if !strings.Contains(out, "doc-sync-reminder") {
		t.Error("output must mention doc-sync-reminder")
	}
}

// Scenario: Backlog section is absent from centinela roadmap output when Backlog phase is missing
func TestDfrc_NoBacklogSectionWhenMissing(t *testing.T) {
	bin := buildCent(t)
	dir := dfrcAcceptDir(t, `{"phases":[{"name":"Phase 0","features":[{"name":"f1"}]}]}`)
	out, code := runCent(t, bin, dir, "roadmap")
	if code != 0 {
		t.Fatalf("roadmap exit=%d\n%s", code, out)
	}
	if strings.Contains(strings.ToLower(out), "backlog") {
		t.Errorf("output must not contain Backlog section when none exists: %s", out)
	}
}

// Scenario: Backlog section is absent when Backlog phase exists but contains no entries
func TestDfrc_NoBacklogSectionWhenEmpty(t *testing.T) {
	bin := buildCent(t)
	dir := dfrcAcceptDir(t, `{"phases":[{"name":"Phase 0","features":[]},{"name":"Backlog","features":[]}]}`)
	out, code := runCent(t, bin, dir, "roadmap")
	if code != 0 {
		t.Fatalf("roadmap exit=%d\n%s", code, out)
	}
	if strings.Contains(strings.ToLower(out), "backlog") {
		t.Errorf("output must not contain Backlog section when empty: %s", out)
	}
}

// Scenario: Backlog features do not appear in centinela roadmap ready output
func TestDfrc_BacklogNotInReadyOutput(t *testing.T) {
	bin := buildCent(t)
	src := `{"phases":[{"name":"Backlog","features":[{"name":"backlog-finding","summary":"s","deferredAt":"t"}]}]}`
	dir := dfrcAcceptDir(t, src)
	out, code := runCent(t, bin, dir, "roadmap", "ready")
	if code != 0 {
		t.Fatalf("roadmap ready exit=%d\n%s", code, out)
	}
	if strings.Contains(out, "backlog-finding") {
		t.Errorf("backlog-finding must not appear in ready output: %s", out)
	}
}

// Scenario: centinela start refuses a Backlog feature with a promote-first error
func TestDfrc_StartRefusesBacklogFeature(t *testing.T) {
	bin := buildCent(t)
	src := `{"phases":[{"name":"Phase 0","features":[{"name":"setup"}]},{"name":"Backlog","features":[{"name":"backlog-finding","summary":"s","deferredAt":"t"}]}]}`
	dir := greenfieldDir(t, src, []string{"backlog-finding"}, true)
	out, code := runCent(t, bin, dir, "start", "backlog-finding")
	if code == 0 {
		t.Fatalf("start backlog-finding must be rejected, got exit 0\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "promot") && !strings.Contains(strings.ToLower(out), "backlog") {
		t.Errorf("output must mention promote: %s", out)
	}
}
