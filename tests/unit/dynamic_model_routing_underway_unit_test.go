package unit_test

import (
	"os"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// dmrUnderwayRepo chdirs into a temp project holding a saved workflow at step
// (RoleScheduledStep resolves the contract from the state file on disk).
func dmrUnderwayRepo(t *testing.T, step string) *workflow.Workflow {
	t.Helper()
	dir := t.TempDir()
	origin, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origin) })   //nolint:errcheck
	os.Chdir(dir)                            //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0o755) //nolint:errcheck
	wf := workflow.New("f")
	wf.CurrentStep = step
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save: %v", err)
	}
	return wf
}

// E6 — a stub left over from an earlier run of the same slug (a re-start, or a
// rewind that never cleaned up) must not close the sanctioned routing window at
// step 1: nothing has been delegated in THIS run.
func TestRoleStepUnderway_StaleEvidenceFromPriorRun(t *testing.T) {
	wf := dmrUnderwayRepo(t, "code")
	path := orchestration.MarkdownPath("f", orchestration.RoleSeniorEngineer)
	if err := os.WriteFile(path, []byte("leftover"), 0o644); err != nil {
		t.Fatal(err)
	}
	stale := wf.StartedAt.Add(-2 * time.Hour)
	if err := os.Chtimes(path, stale, stale); err != nil {
		t.Fatal(err)
	}
	if workflow.RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("a stub predating startedAt is leftover, not evidence that delegation began")
	}
	// Rewritten during this run, it counts again.
	now := wf.StartedAt.Add(time.Minute)
	if err := os.Chtimes(path, now, now); err != nil {
		t.Fatal(err)
	}
	if !workflow.RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("an artifact written since startedAt means delegation began")
	}
}

// E6 (json half) — evidence init writes both stubs; either one, from this run,
// proves delegation began.
func TestRoleStepUnderway_JSONStubFromThisRunCounts(t *testing.T) {
	wf := dmrUnderwayRepo(t, "code")
	path := orchestration.JSONPath("f", orchestration.RoleSeniorEngineer)
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !workflow.RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("the JSON stub must count as well as the markdown one")
	}
}

// E7 — a currentStep off the step order made stepIndexIn return -1 for BOTH
// operands, so `scheduled < current` was never true and every underway refusal
// was disarmed at once. The fail-safe reading is "underway".
func TestRoleStepUnderway_OutOfOrderCurrentStep(t *testing.T) {
	wf := dmrUnderwayRepo(t, "docs")
	wf.CurrentStep = "foo"
	for _, role := range []orchestration.Role{
		orchestration.RoleSeniorEngineer, orchestration.RolePlanner, orchestration.RoleGatekeeper,
	} {
		if !workflow.RoleStepUnderway(wf, role) {
			t.Fatalf("an unreadable cursor must not free %s to be downgraded", role)
		}
	}
	// An unscheduled role stays unscheduled — the guard does not invent steps.
	if workflow.RoleStepUnderway(wf, orchestration.RoleMergeSteward) {
		t.Fatal("an unscheduled role is never underway")
	}
}
