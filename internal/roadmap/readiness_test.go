package roadmap

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/workflow"
)

func seedStep(t *testing.T, name, step string) {
	t.Helper()
	wf := &workflow.Workflow{Feature: name, CurrentStep: step, Steps: map[string]workflow.StepState{}}
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("seed %s: %v", name, err)
	}
}

func rdMap(deps map[string][]string, names ...string) *Roadmap {
	p := Phase{Name: "P"}
	for _, n := range names {
		p.Features = append(p.Features, Feature{Name: n, DependsOn: deps[n]})
	}
	return &Roadmap{Phases: []Phase{p}}
}

func find(rs []FeatureReadiness, name string) FeatureReadiness {
	for _, fr := range rs {
		if fr.Name == name {
			return fr
		}
	}
	return FeatureReadiness{}
}

// DeriveReadiness/classifyFeature/collectUnmet: each of the four states plus an
// in-progress dep that keeps the dependent blocked and multi-dep BlockedBy.
func TestDeriveReadiness_States(t *testing.T) {
	chdirRoadmapTemp(t)
	seedDone(t, "d")
	seedStep(t, "w", "code")
	r := rdMap(map[string][]string{
		"ready": {"d"},
		"blk":   {"w", "p"},
	}, "d", "w", "p", "ready", "blk")
	got := DeriveReadiness(r)
	if find(got, "d").State != "done" {
		t.Fatalf("d should be done: %+v", find(got, "d"))
	}
	if find(got, "w").State != "in-progress" {
		t.Fatalf("w should be in-progress: %+v", find(got, "w"))
	}
	if find(got, "ready").State != "ready" {
		t.Fatalf("ready should be ready: %+v", find(got, "ready"))
	}
	blk := find(got, "blk")
	if blk.State != "blocked" {
		t.Fatalf("blk should be blocked: %+v", blk)
	}
	joined := strings.Join(blk.BlockedBy, ",")
	if !strings.Contains(joined, "w") || !strings.Contains(joined, "p") {
		t.Fatalf("blk must be blocked by in-progress + planned dep: %v", blk.BlockedBy)
	}
	if DeriveReadiness(nil) != nil {
		t.Fatal("nil roadmap should derive nil")
	}
}

// ReadySet returns ready names in declared order across the roadmap.
func TestReadySet_DeclaredOrder(t *testing.T) {
	chdirRoadmapTemp(t)
	seedDone(t, "a")
	r := rdMap(map[string][]string{"b": {"a"}, "c": {"a", "z"}}, "a", "b", "c", "z")
	if got := ReadySet(r); strings.Join(got, ",") != "b,z" {
		t.Fatalf("ReadySet want [b z], got %v", got)
	}
}

// UnmetDependencies returns unmet dep names; nil for satisfied/none/missing/nil.
func TestUnmetDependencies(t *testing.T) {
	chdirRoadmapTemp(t)
	seedDone(t, "a")
	r := rdMap(map[string][]string{"b": {"a"}, "c": {"a", "z"}}, "a", "b", "c", "z")
	if um := UnmetDependencies(r, "c"); strings.Join(um, ",") != "z" {
		t.Fatalf("UnmetDependencies(c) want [z], got %v", um)
	}
	if um := UnmetDependencies(r, "b"); um != nil {
		t.Fatalf("UnmetDependencies(b) want nil, got %v", um)
	}
	if um := UnmetDependencies(r, "missing"); um != nil {
		t.Fatalf("UnmetDependencies(missing) want nil, got %v", um)
	}
	if um := UnmetDependencies(nil, "b"); um != nil {
		t.Fatalf("UnmetDependencies(nil) want nil, got %v", um)
	}
}
