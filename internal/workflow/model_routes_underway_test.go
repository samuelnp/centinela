package workflow

import (
	"os"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// underwayRepo chdirs into a temp project holding a saved workflow at step.
func underwayRepo(t *testing.T, step string) *Workflow {
	t.Helper()
	dir := t.TempDir()
	origin, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origin) }) //nolint:errcheck
	os.Chdir(dir)                          //nolint:errcheck
	os.MkdirAll(WorkflowDir, 0755)         //nolint:errcheck
	wf := New("f")
	wf.CurrentStep = step
	if err := Save(wf); err != nil {
		t.Fatalf("save: %v", err)
	}
	return wf
}

func TestRoleStepUnderway_CurrentStepNeedsEvidenceOnDisk(t *testing.T) {
	wf := underwayRepo(t, "code")
	if RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("the sanctioned routing window is the current step BEFORE evidence exists")
	}
	path := orchestration.MarkdownPath("f", orchestration.RoleSeniorEngineer)
	os.WriteFile(path, []byte("stub"), 0644) //nolint:errcheck
	if !RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("an existing evidence artifact means delegation began")
	}
}

func TestRoleStepUnderway_PastCurrentDoneAndUnscheduled(t *testing.T) {
	wf := underwayRepo(t, "tests")
	if !RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("a step behind the cursor is underway regardless of evidence")
	}
	if RoleStepUnderway(wf, orchestration.RoleGatekeeper) {
		t.Fatal("a step ahead of the cursor is not underway")
	}
	if RoleStepUnderway(wf, orchestration.RoleMergeSteward) {
		t.Fatal("an unscheduled role is never underway")
	}
	wf.CurrentStep = "done"
	if !RoleStepUnderway(wf, orchestration.RoleGatekeeper) {
		t.Fatal("a completed workflow leaves every step underway")
	}
}

// A stub predating StartedAt is left over from an earlier run of the same slug
// and must not close the sanctioned start-time routing window.
func TestRoleStepUnderway_StaleEvidencePredatingStartedAt(t *testing.T) {
	wf := underwayRepo(t, "code")
	path := orchestration.JSONPath("f", orchestration.RoleSeniorEngineer)
	os.WriteFile(path, []byte("{}"), 0644) //nolint:errcheck
	stale := wf.StartedAt.Add(-time.Hour)
	os.Chtimes(path, stale, stale) //nolint:errcheck
	if RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("a leftover stub is not evidence that THIS run delegated")
	}
	fresh := wf.StartedAt.Add(time.Minute)
	os.Chtimes(path, fresh, fresh) //nolint:errcheck
	if !RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("an artifact written since startedAt means delegation began")
	}
}

// A cursor off the step order proves nothing; the fail-safe reading is underway.
func TestRoleStepUnderway_CurrentStepOutsideTheOrder(t *testing.T) {
	wf := underwayRepo(t, "code")
	wf.CurrentStep = "foo"
	if !RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("an unreadable cursor must not disarm the downgrade refusal")
	}
	if RoleStepUnderway(wf, orchestration.RoleMergeSteward) {
		t.Fatal("an unscheduled role is still never underway")
	}
}
