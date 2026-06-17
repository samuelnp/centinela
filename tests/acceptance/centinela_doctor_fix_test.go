package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-doctor.feature

// blockRoadmapWrite makes roadmap.json + its parent read-only so the safe
// glyph-strip repair's rewrite fails at runtime (forces a partial failure).
func blockRoadmapWrite(t *testing.T, dir string) {
	t.Helper()
	rj := filepath.Join(dir, ".workflow", "roadmap.json")
	if err := os.Chmod(rj, 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(dir, ".workflow"), 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(filepath.Join(dir, ".workflow"), 0o755)
		_ = os.Chmod(rj, 0o644)
	})
}

// Scenario: --fix attempts all safe repairs even when one fails
func TestDoctorFixAttemptsAllEvenWhenOneFails(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	// fixable #1: missing hooks. fixable #3: orphaned tmp.
	writeFile(t, dir, ".claude/settings.json", "{}")
	writeFile(t, dir, ".workflow/feat-qa-senior.json.tmp", "{}")
	// fixable #2 (made to fail): glyph repair blocked by a read-only roadmap.json.
	seedRoadmap(t, dir, "✅ Phase 0: Bootstrap")
	blockRoadmapWrite(t, dir)
	out, code := runDoctor(t, dir, "--fix")
	if code != 1 {
		t.Fatalf("a failed repair must drive exit 1, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "✓ hooks") {
		t.Fatalf("hooks (first) must still repair:\n%s", out)
	}
	if !strings.Contains(out, "✗ roadmap") {
		t.Fatalf("failed roadmap repair must render ✗:\n%s", out)
	}
}

// Scenario: --fix partial success renders a clear per-check post-fix report
func TestDoctorFixPartialReportPerCheck(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, ".claude/settings.json", "{}")
	seedRoadmap(t, dir, "✅ Phase 0: Bootstrap")
	blockRoadmapWrite(t, dir)
	out, _ := runDoctor(t, dir, "--fix")
	if !strings.Contains(out, "✓ hooks") || !strings.Contains(out, "✗ roadmap") {
		t.Fatalf("each check line must reflect its own post-fix state:\n%s", out)
	}
	// summary must reflect at least one error.
	if !strings.Contains(lastLine(out), "1 error") && !strings.Contains(lastLine(out), "error") {
		t.Fatalf("summary must reflect post-fix results: %q", lastLine(out))
	}
}

// Scenario: --fix never performs destructive actions — worktree and .workflow intact after fix
func TestDoctorFixNeverDestructive(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	abandonedWorktree(t, dir, "gone")
	writeFile(t, dir, ".workflow/ghost.json",
		`{"feature":"ghost","currentStep":"code","steps":{}}`)
	writeFile(t, dir, ".workflow/feat-qa-senior.json.tmp", "{}")
	out, _ := runDoctor(t, dir, "--fix")
	if _, err := os.Stat(filepath.Join(dir, ".worktrees", "gone")); err != nil {
		t.Fatal("worktree must survive --fix")
	}
	if _, err := os.Stat(filepath.Join(dir, ".workflow", "ghost.json")); err != nil {
		t.Fatal("workflow state must survive --fix")
	}
	if left, _ := filepath.Glob(filepath.Join(dir, ".workflow", "*.json.tmp")); len(left) != 0 {
		t.Fatalf("safe tmp sweep should have run, left %v", left)
	}
	if !strings.Contains(out, "worktrees") || !strings.Contains(out, "workflow-state") {
		t.Fatalf("destructive findings must remain with manual commands:\n%s", out)
	}
}
