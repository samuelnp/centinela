package teamdashboard

import (
	"reflect"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/telemetry"
	"github.com/samuelnp/centinela/internal/workflow"
)

func sampleInputs() Inputs {
	now := time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC)
	return Inputs{
		Active: []*workflow.Workflow{
			wf("alpha", "code", now.Add(-48*time.Hour)),
			wf("beta", "tests", now.Add(-24*time.Hour)),
		},
		Roadmap: &roadmap.Roadmap{Phases: []roadmap.Phase{phase("Q1", "f1", "f2")}},
		Events:  []telemetry.Event{gateEvent("coverage"), gateEvent("coverage")},
		Owners:  map[string]string{"alpha": "Alice"},
		Now:     now,
	}
}

func TestCompute_AssemblesAllThreePanels(t *testing.T) {
	d := Compute(sampleInputs())
	if len(d.Features) != 2 {
		t.Fatalf("features: %+v", d.Features)
	}
	if !d.Roadmap.Present || d.Roadmap.Total != 2 {
		t.Fatalf("roadmap: %+v", d.Roadmap)
	}
	if len(d.Gates) != 1 || d.Gates[0].Gate != "coverage" || d.Gates[0].Fails != 2 {
		t.Fatalf("gates: %+v", d.Gates)
	}
}

func TestCompute_EmptyInputsHonestEmptyState(t *testing.T) {
	d := Compute(Inputs{})
	if len(d.Features) != 0 {
		t.Fatalf("no active => empty features, got %+v", d.Features)
	}
	if d.Roadmap.Present {
		t.Fatalf("nil roadmap => Present:false, got %+v", d.Roadmap)
	}
	if len(d.Gates) != 0 {
		t.Fatalf("no events => empty gates, got %+v", d.Gates)
	}
}

func TestCompute_DeterministicSameInputs(t *testing.T) {
	in := sampleInputs()
	if !reflect.DeepEqual(Compute(in), Compute(in)) {
		t.Fatal("Compute must be deterministic for identical Inputs")
	}
}
