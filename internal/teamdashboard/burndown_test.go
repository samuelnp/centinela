package teamdashboard

import (
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func phase(name string, features ...string) roadmap.Phase {
	p := roadmap.Phase{Name: name}
	for _, f := range features {
		p.Features = append(p.Features, roadmap.Feature{Name: f})
	}
	return p
}

func TestBurndown_NilRoadmapEmptyState(t *testing.T) {
	b := burndown(nil)
	if b.Present {
		t.Fatalf("nil roadmap must be Present:false, got %+v", b)
	}
	if b.Total != 0 || len(b.Phases) != 0 {
		t.Fatalf("nil roadmap must be empty, got %+v", b)
	}
}

func TestBurndown_EmptyRoadmapZeroTotals(t *testing.T) {
	b := burndown(&roadmap.Roadmap{})
	if !b.Present {
		t.Fatal("present roadmap must be Present:true")
	}
	if b.Total != 0 || len(b.Phases) != 0 {
		t.Fatalf("empty roadmap => 0/0, got %+v", b)
	}
}

func TestBurndown_ExcludesBacklogBaselineCountsSchedulable(t *testing.T) {
	// No workflow files on disk here, so every FeatureStatus resolves to
	// "planned" (done=0). We assert schedulable totals + Backlog/Baseline skip.
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{
		phase("Backlog", "deferred-1"),
		phase("Q1", "f1", "f2"),
		phase("Baseline", "existing-1"),
		phase("Q2", "f3"),
	}}
	b := burndown(r)
	if !b.Present {
		t.Fatal("present")
	}
	if b.Total != 3 || b.Planned != 3 || b.Done != 0 {
		t.Fatalf("schedulable totals wrong: %+v", b)
	}
	if len(b.Phases) != 2 {
		t.Fatalf("Backlog/Baseline must be excluded, got phases %+v", b.Phases)
	}
	if b.Phases[0].Name != "Q1" || b.Phases[0].Total != 2 {
		t.Fatalf("Q1 phase: %+v", b.Phases[0])
	}
	if b.Phases[1].Name != "Q2" || b.Phases[1].Total != 1 {
		t.Fatalf("Q2 phase: %+v", b.Phases[1])
	}
}
