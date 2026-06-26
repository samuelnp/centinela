package teamdashboard

import (
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/workflow"
)

func wf(feature, step string, started time.Time) *workflow.Workflow {
	return &workflow.Workflow{
		Feature:            feature,
		CurrentStep:        step,
		StartedAt:          started,
		StepOrder:          []string{"plan", "code", "tests", "validate", "docs"},
		EnforcementProfile: "strict",
		Archetype:          "hexagonal",
		WorktreePath:       ".worktrees/" + feature,
	}
}

func TestFeatures_RowFieldsAndOwner(t *testing.T) {
	now := time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC)
	started := now.Add(-72 * time.Hour) // 3 days
	owners := map[string]string{"alpha": "Alice"}
	rows := features([]*workflow.Workflow{wf("alpha", "tests", started)}, owners, now)
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.Feature != "alpha" || r.Step != "tests" {
		t.Fatalf("feature/step: %+v", r)
	}
	if r.StepIndex != 2 || r.StepTotal != 5 {
		t.Fatalf("step index/total: got %d/%d want 2/5", r.StepIndex, r.StepTotal)
	}
	if r.AgeDays != 3 {
		t.Fatalf("age: got %d want 3", r.AgeDays)
	}
	if r.Profile != "strict" || r.Archetype != "hexagonal" || r.Worktree != ".worktrees/alpha" {
		t.Fatalf("passthrough: %+v", r)
	}
	if r.Owner != "Alice" {
		t.Fatalf("owner: got %q want Alice", r.Owner)
	}
}

func TestFeatures_OwnerUnknownAndNilSkipAndOrder(t *testing.T) {
	now := time.Now().UTC()
	in := []*workflow.Workflow{wf("a", "code", now), nil, wf("b", "plan", now)}
	rows := features(in, map[string]string{"a": ""}, now)
	if len(rows) != 2 {
		t.Fatalf("nil should be skipped: got %d", len(rows))
	}
	if rows[0].Feature != "a" || rows[1].Feature != "b" {
		t.Fatalf("order not preserved: %+v", rows)
	}
	if rows[0].Owner != "unknown" || rows[1].Owner != "unknown" {
		t.Fatalf("empty/missing owner should be unknown: %+v", rows)
	}
}
