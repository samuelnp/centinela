package unit_test

import (
	"os"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
	"github.com/samuelnp/centinela/internal/workflow"
)

// skewStamp writes the artifact and forces its mtime, reproducing a kernel
// stamping it from a coarse tick-granularity clock that reads slightly earlier
// than the precise clock StartedAt came from. Forcing it is the point: the
// original tests inferred the ordering from the host clock, so they could not
// fail on macOS and the defect reached CI unseen.
func skewStamp(t *testing.T, path string, at time.Time) {
	t.Helper()
	if err := os.WriteFile(path, []byte("stub"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	if err := os.Chtimes(path, at, at); err != nil {
		t.Fatalf("chtimes %s: %v", path, err)
	}
}

// Regression for the CI-only failure: an artifact stamped a few milliseconds
// before StartedAt was read as left over from an earlier run, silencing the
// underway refusal for a window right after `start`.
func TestRoleStepUnderway_CoarseClockSkewStillCounts(t *testing.T) {
	wf := dmrUnderwayRepo(t, "code")
	path := orchestration.JSONPath("f", orchestration.RoleSeniorEngineer)
	skewStamp(t, path, wf.StartedAt.Add(-8*time.Millisecond))

	if !workflow.RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("an artifact stamped by a coarse clock must count as from this run")
	}
}

// The tolerance must stay narrow enough to keep doing its job.
func TestRoleStepUnderway_SkewToleranceDoesNotAdoptOldStubs(t *testing.T) {
	wf := dmrUnderwayRepo(t, "code")
	path := orchestration.JSONPath("f", orchestration.RoleSeniorEngineer)
	skewStamp(t, path, wf.StartedAt.Add(-10*time.Minute))

	if workflow.RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("a stub predating the run by minutes must not close the routing window")
	}
}
