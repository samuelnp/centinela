// Acceptance: specs/durable-workflow-state.feature
package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dwsTemps returns the atomic-write temp files left beside the state files.
func dwsTemps(t *testing.T, dir string) []string {
	t.Helper()
	m, _ := filepath.Glob(filepath.Join(dir, ".workflow", ".*.json.tmp-*"))
	return m
}

// Scenario: A completed write replaces the state file in one step
func TestAccCompletedWriteLeavesNoTemp(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "code", ""))

	dmrOK(t, bin, dir, "route", "set", "f", "senior-engineer", "balanced",
		"--reason", "config-only change")
	if got := dmrState(t, dir, "f"); !strings.Contains(got, "senior-engineer") {
		t.Fatalf("the write did not land:\n%s", got)
	}
	if left := dwsTemps(t, dir); len(left) != 0 {
		t.Fatalf("temporary files remain beside the state file: %v", left)
	}
}

// Scenario: The replaced state file keeps its readable file mode
func TestAccReplacedStateFileKeepsItsMode(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "code", ""))
	path := filepath.Join(dir, ".workflow", "f.json")
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}

	dmrOK(t, bin, dir, "route", "set", "f", "senior-engineer", "balanced",
		"--reason", "config-only change")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("mode = %v, want 0644 — every hook reads this file", info.Mode().Perm())
	}
}

// Scenario: A write that cannot be completed reports the state file path
func TestAccUnwritableStateDirReportsThePath(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "code", ""))
	before := dmrState(t, dir, "f")
	wfDir := filepath.Join(dir, ".workflow")
	if err := os.Chmod(wfDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(wfDir, 0o755) }) //nolint:errcheck

	out := dmrRefused(t, bin, dir, "route", "set", "f", "senior-engineer",
		"balanced", "--reason", "config-only change")
	if !strings.Contains(out, ".workflow/f.json") {
		t.Fatalf("the error must name the state file, got:\n%s", out)
	}
	if got := dmrState(t, dir, "f"); got != before {
		t.Fatal("a failed write must leave the state file untouched")
	}
	if left := dwsTemps(t, dir); len(left) != 0 {
		t.Fatalf("a failed write must leave no partial file: %v", left)
	}
}

// Scenario: A same-version file round-trips without losing fields
func TestAccSameVersionRoundTripKeepsEveryField(t *testing.T) {
	bin := dmrBuildBin(t)
	dir := dmrRepo(t, dmrDynamicTOML)
	dmrWrite(t, dir, "f", dmrWorkflowJSON("f", "code", ""))
	dmrOK(t, bin, dir, "route", "set", "f", "senior-engineer", "balanced",
		"--reason", "config-only change")
	before := dmrState(t, dir, "f")

	dmrOK(t, bin, dir, "route", "set", "f", "qa-senior", "balanced",
		"--reason", "config-only change")
	after := dmrState(t, dir, "f")
	for _, want := range []string{"senior-engineer", "qa-senior", dmrStartedAt,
		"strict-subagents-v1", "planner-v1", "adversarial-v1", `"schemaVersion": 1`} {
		if !strings.Contains(after, want) {
			t.Fatalf("the round-trip lost %q\nbefore:\n%s\nafter:\n%s", want, before, after)
		}
	}
}
