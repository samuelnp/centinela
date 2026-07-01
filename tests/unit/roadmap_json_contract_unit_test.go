package unit_test

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

// BuildView projects persisted features into status/readiness/counts, deriving
// readiness only for planned rows and excluding non-schedulable phases.
func TestBuildViewProjectsStatusReadinessAndCounts(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                        //nolint:errcheck
	os.Chdir(d)                              //nolint:errcheck
	os.MkdirAll(workflow.WorkflowDir, 0o755) //nolint:errcheck
	done := workflow.New("auth-service")
	done.CurrentStep = "done"
	if err := workflow.Save(done); err != nil {
		t.Fatalf("seed done: %v", err)
	}
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{
		{Name: "Backlog", Features: []roadmap.Feature{{Name: "deferred"}}},
		{Name: "Q1", Features: []roadmap.Feature{
			{Name: "auth-service"},
			{Name: "checkout-ui", DependsOn: []string{"auth-service"}},
			{Name: "reporting", DependsOn: []string{"missing"}},
		}},
	}}
	v := roadmap.BuildView(r)
	if len(v.Phases) != 1 || v.Phases[0].Name != "Q1" {
		t.Fatalf("Backlog must be excluded, got %+v", v.Phases)
	}
	if v.Counts != (roadmap.StatusCounts{Planned: 2, Done: 1}) {
		t.Fatalf("counts = %+v", v.Counts)
	}
	f := v.Phases[0].Features
	if f[0].Status != "done" || f[0].Readiness != "" {
		t.Fatalf("done row must omit readiness: %+v", f[0])
	}
	if f[1].Readiness != "ready" || f[1].BlockedBy != nil {
		t.Fatalf("checkout-ui must be ready with no blockedBy: %+v", f[1])
	}
	if f[2].Readiness != "blocked" || len(f[2].BlockedBy) != 1 || f[2].BlockedBy[0] != "missing" {
		t.Fatalf("reporting must be blocked by its unmet dep: %+v", f[2])
	}
}
