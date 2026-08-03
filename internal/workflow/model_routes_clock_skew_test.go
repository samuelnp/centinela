package workflow

import (
	"os"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/orchestration"
)

// stampMtime forces an artifact's mtime, reproducing a coarse-clock kernel
// stamping it slightly before the precise time StartedAt was taken. Doing it
// explicitly is the point: the original tests depended on the host's timestamp
// granularity, so they could not fail on macOS and only surfaced in CI.
func stampMtime(t *testing.T, path string, at time.Time) {
	t.Helper()
	if err := os.WriteFile(path, []byte("stub"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	if err := os.Chtimes(path, at, at); err != nil {
		t.Fatalf("chtimes %s: %v", path, err)
	}
}

// Regression: an artifact whose mtime lands a few milliseconds BEFORE StartedAt
// was misjudged as left over from an earlier run, so the underway refusal went
// silent for a window right after `start`.
func TestRoleEvidenceFromThisRun_ToleratesCoarseMtimeSkew(t *testing.T) {
	wf := underwayRepo(t, "code")
	path := orchestration.MarkdownPath("f", orchestration.RoleSeniorEngineer)
	stampMtime(t, path, wf.StartedAt.Add(-8*time.Millisecond))

	if !RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("an artifact stamped by a coarse clock must still count as from this run")
	}
}

// The tolerance must not swallow what the check exists for: a stub from an
// earlier run of the same slug still leaves the routing window open.
func TestRoleEvidenceFromThisRun_StillRejectsAnEarlierRunsStub(t *testing.T) {
	wf := underwayRepo(t, "code")
	path := orchestration.MarkdownPath("f", orchestration.RoleSeniorEngineer)
	stampMtime(t, path, wf.StartedAt.Add(-30*time.Minute))

	if RoleStepUnderway(wf, orchestration.RoleSeniorEngineer) {
		t.Fatal("a stub predating the run by minutes must not close the routing window")
	}
}

// The boundary itself: just outside the grace window is still an earlier run.
func TestRoleEvidenceFromThisRun_GraceWindowBoundary(t *testing.T) {
	path := orchestration.MarkdownPath("f", orchestration.RoleSeniorEngineer)

	inside := underwayRepo(t, "code")
	stampMtime(t, path, inside.StartedAt.Add(-clockSkewGrace+time.Second))
	if !RoleStepUnderway(inside, orchestration.RoleSeniorEngineer) {
		t.Fatal("inside the grace window must count as from this run")
	}

	outside := underwayRepo(t, "code")
	stampMtime(t, path, outside.StartedAt.Add(-clockSkewGrace-time.Second))
	if RoleStepUnderway(outside, orchestration.RoleSeniorEngineer) {
		t.Fatal("outside the grace window must not count")
	}
}
