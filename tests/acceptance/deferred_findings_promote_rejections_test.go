package acceptance_test

// Acceptance: specs/deferred-findings-roadmap-capture.feature

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Scenario: Promote with a low overall score is recorded rather than rejected
func TestDfrc_PromoteLowOverallRecorded(t *testing.T) {
	bin := buildCent(t)
	dir := dfrcAcceptDir(t, promoteRoadmap)
	seedPromoteArtifacts(t, dir)
	out, code := runCent(t, bin, dir, "roadmap", "promote", "hook-timeout-config",
		"--phase", "Phase 5 — Operability & DX", "--scores", "9,9,8,7,9,3")
	if code != 0 {
		t.Fatalf("a low overall score must no longer be rejected, got exit %d\n%s", code, out)
	}
	quality, err := os.ReadFile(filepath.Join(dir, ".workflow", "roadmap-quality.json"))
	if err != nil {
		t.Fatalf("read quality: %v", err)
	}
	if !strings.Contains(string(quality), `"overall":3`) {
		t.Errorf("the low score must be RECORDED verbatim: %s", quality)
	}
}

// Scenario: Promote with any dimension score outside 1-10 is rejected before any write
func TestDfrc_PromoteOutOfRangeScoreRejected(t *testing.T) {
	bin := buildCent(t)
	dir := dfrcAcceptDir(t, promoteRoadmap)
	seedPromoteArtifacts(t, dir)
	out, code := runCent(t, bin, dir, "roadmap", "promote", "hook-timeout-config",
		"--phase", "Phase 5 — Operability & DX", "--scores", "11,9,9,9,9,9")
	if code == 0 {
		t.Fatalf("out-of-range score must be rejected, got exit 0\n%s", out)
	}
	_ = out
}

// Scenario: Promote into a non-existent phase is rejected with known phases listed
func TestDfrc_PromoteUnknownPhaseRejected(t *testing.T) {
	bin := buildCent(t)
	src := `{"phases":[{"name":"Phase 0: Bootstrap","features":[]},{"name":"Phase 5 — Operability & DX","features":[]},{"name":"Backlog","features":[{"name":"phase-test","summary":"s","deferredAt":"t"}]}]}`
	dir := dfrcAcceptDir(t, src)
	seedPromoteArtifacts(t, dir)
	out, code := runCent(t, bin, dir, "roadmap", "promote", "phase-test",
		"--phase", "Phase 99 — Does Not Exist", "--scores", "9,9,9,9,9,9")
	if code == 0 {
		t.Fatalf("unknown phase must be rejected, got exit 0\n%s", out)
	}
	if !strings.Contains(out, "Phase 0") && !strings.Contains(out, "Phase 5") {
		t.Errorf("output must list known phases: %s", out)
	}
}

// Scenario: Promote a slug not in the Backlog phase is rejected cleanly
func TestDfrc_PromoteSlugNotInBacklog(t *testing.T) {
	bin := buildCent(t)
	src := `{"phases":[{"name":"Phase 5","features":[]},{"name":"Backlog","features":[]}]}`
	dir := dfrcAcceptDir(t, src)
	seedPromoteArtifacts(t, dir)
	out, code := runCent(t, bin, dir, "roadmap", "promote", "not-in-backlog",
		"--phase", "Phase 5", "--scores", "9,9,9,9,9,9")
	if code == 0 {
		t.Fatalf("slug not in Backlog must be rejected, got exit 0\n%s", out)
	}
	_ = out
}

// Scenario: Promote with a malformed --scores CSV is rejected before any write
func TestDfrc_PromoteMalformedScoresRejected(t *testing.T) {
	bin := buildCent(t)
	dir := dfrcAcceptDir(t, promoteRoadmap)
	seedPromoteArtifacts(t, dir)
	out, code := runCent(t, bin, dir, "roadmap", "promote", "hook-timeout-config",
		"--phase", "Phase 5 — Operability & DX", "--scores", "9,9,9")
	if code == 0 {
		t.Fatalf("malformed scores must be rejected, got exit 0\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "six") && !strings.Contains(strings.ToLower(out), "6") && !strings.Contains(strings.ToLower(out), "scores") {
		t.Errorf("output must mention six scores requirement: %s", out)
	}
}

// Scenario: --scores "" (empty string) is treated as a usage error (regression)
func TestDfrc_PromoteEmptyScoresError(t *testing.T) {
	bin := buildCent(t)
	dir := dfrcAcceptDir(t, promoteRoadmap)
	seedPromoteArtifacts(t, dir)
	out, code := runCent(t, bin, dir, "roadmap", "promote", "hook-timeout-config",
		"--phase", "Phase 5 — Operability & DX", "--scores", "")
	if code == 0 {
		t.Fatalf("empty --scores must be rejected, got exit 0\n%s", out)
	}
}
