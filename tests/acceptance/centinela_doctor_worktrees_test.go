package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/centinela-doctor.feature

// abandonedWorktree creates a .worktrees/<feat> dir with no branch (a missing
// branch is treated as merged → abandoned).
func abandonedWorktree(t *testing.T, dir, feat string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".worktrees", feat), 0o755); err != nil {
		t.Fatal(err)
	}
}

// Scenario: Abandoned worktree for a merged branch is reported with the removal command
func TestDoctorAbandonedWorktreeReported(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	abandonedWorktree(t, dir, "gone")
	out, code := runDoctor(t, dir)
	if code != 1 || !strings.Contains(out, "✗ worktrees") {
		t.Fatalf("abandoned worktree must Error/exit 1, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "git worktree remove") {
		t.Fatalf("must surface the remove command:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".worktrees", "gone")); err != nil {
		t.Fatal("worktree dir must still exist (read-only)")
	}
}

// Scenario: --fix does NOT remove an abandoned worktree
func TestDoctorFixDoesNotRemoveWorktree(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	abandonedWorktree(t, dir, "gone")
	out, _ := runDoctor(t, dir, "--fix")
	if _, err := os.Stat(filepath.Join(dir, ".worktrees", "gone")); err != nil {
		t.Fatal("--fix must NOT remove the worktree")
	}
	if strings.Contains(out, "✓ worktrees") {
		t.Fatalf("worktrees finding must remain after --fix:\n%s", out)
	}
	if !strings.Contains(out, "git worktree remove") {
		t.Fatalf("removal command must still be shown:\n%s", out)
	}
}

// Scenario: No worktrees present causes worktrees check to report OK
func TestDoctorNoWorktreesOK(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	out, _ := runDoctor(t, dir)
	if !strings.Contains(out, "✓ worktrees") {
		t.Fatalf("no worktrees must be OK:\n%s", out)
	}
}

// Scenario: Orphaned .workflow state with no corresponding branch is reported
func TestDoctorOrphanedWorkflowStateReported(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, ".workflow/ghost.json",
		`{"feature":"ghost","currentStep":"code","steps":{}}`)
	out, code := runDoctor(t, dir)
	if code != 1 || !strings.Contains(out, "✗ workflow-state") {
		t.Fatalf("orphan must Error/exit 1, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "ghost") || !strings.Contains(out, "rm ") {
		t.Fatalf("must name file + manual rm command:\n%s", out)
	}
}

// Scenario: --fix does NOT delete orphaned .workflow state
func TestDoctorFixDoesNotDeleteWorkflowState(t *testing.T) {
	dir := doctorRepo(t)
	gitInit(t, dir)
	writeFile(t, dir, ".workflow/ghost.json",
		`{"feature":"ghost","currentStep":"code","steps":{}}`)
	out, _ := runDoctor(t, dir, "--fix")
	if _, err := os.Stat(filepath.Join(dir, ".workflow", "ghost.json")); err != nil {
		t.Fatal("--fix must NOT delete workflow state")
	}
	if !strings.Contains(out, "workflow-state") || !strings.Contains(out, "rm ") {
		t.Fatalf("finding + manual command must persist:\n%s", out)
	}
}
