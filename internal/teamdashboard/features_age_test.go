package teamdashboard

import (
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/workflow"
)

func TestAgeDays_ZeroFutureNormalFloor(t *testing.T) {
	now := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		started time.Time
		want    int
	}{
		{"zero", time.Time{}, 0},
		{"future", now.Add(48 * time.Hour), 0},
		{"now", now, 0},
		{"floor-7d23h", now.Add(-(7*24 + 23) * time.Hour), 7},
		{"exact-1d", now.Add(-24 * time.Hour), 1},
	}
	for _, c := range cases {
		if got := ageDays(c.started, now); got != c.want {
			t.Fatalf("%s: ageDays = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestDoneCount_DonePositionUnknown(t *testing.T) {
	steps := []string{"plan", "code", "tests", "validate", "docs"}
	mk := func(step string) *workflow.Workflow {
		return &workflow.Workflow{CurrentStep: step, StepOrder: steps}
	}
	if got := doneCount(mk("done")); got != 5 {
		t.Fatalf("done => %d, want 5", got)
	}
	if got := doneCount(mk("validate")); got != 3 {
		t.Fatalf("validate => %d, want 3", got)
	}
	if got := doneCount(mk("plan")); got != 0 {
		t.Fatalf("plan => %d, want 0", got)
	}
	if got := doneCount(mk("mystery")); got != 0 {
		t.Fatalf("unknown step => %d, want 0", got)
	}
}

func TestOwnerOf_PresentEmptyMissing(t *testing.T) {
	owners := map[string]string{"a": "Alice", "b": ""}
	if ownerOf(owners, "a") != "Alice" {
		t.Fatal("present owner")
	}
	if ownerOf(owners, "b") != "unknown" {
		t.Fatal("empty owner => unknown")
	}
	if ownerOf(owners, "missing") != "unknown" {
		t.Fatal("missing owner => unknown")
	}
}
