package unit_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

// rpChdir moves into a fresh temp dir with .workflow/ so FeatureStatus reads from
// per-feature workflow files seeded by rpDone/rpInProgress; absent file == planned.
func rpChdir(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workflow.WorkflowDir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func rpSave(t *testing.T, name, step string) {
	t.Helper()
	if err := workflow.Save(&workflow.Workflow{Feature: name, CurrentStep: step, Steps: map[string]workflow.StepState{}}); err != nil {
		t.Fatalf("save %s: %v", name, err)
	}
}

// rpRoadmap builds a single-phase roadmap; deps maps a feature name to its dependsOn.
func rpRoadmap(deps map[string][]string, features ...string) *roadmap.Roadmap {
	p := roadmap.Phase{Name: "Phase 0"}
	for _, f := range features {
		p.Features = append(p.Features, roadmap.Feature{Name: f, DependsOn: deps[f]})
	}
	return &roadmap.Roadmap{Phases: []roadmap.Phase{p}}
}

func stateOf(rs []roadmap.FeatureReadiness, name string) roadmap.FeatureReadiness {
	for _, fr := range rs {
		if fr.Name == name {
			return fr
		}
	}
	return roadmap.FeatureReadiness{}
}

// DeriveReadiness classifies done / in-progress / ready / blocked; an in-progress
// dependency keeps the dependent blocked and all unmet deps are listed.
func TestDeriveReadiness_FourStatesAndBlockedBy(t *testing.T) {
	rpChdir(t)
	rpSave(t, "done-f", "done")
	rpSave(t, "wip-f", "code")
	r := rpRoadmap(map[string][]string{
		"ready-f": {"done-f"},
		"blk-f":   {"wip-f", "planned-x"},
	}, "done-f", "wip-f", "planned-x", "ready-f", "blk-f")
	got := roadmap.DeriveReadiness(r)
	if s := stateOf(got, "done-f"); s.State != "done" || len(s.BlockedBy) != 0 {
		t.Fatalf("done-f: got %+v", s)
	}
	if s := stateOf(got, "wip-f"); s.State != "in-progress" || len(s.BlockedBy) != 0 {
		t.Fatalf("wip-f: got %+v", s)
	}
	if s := stateOf(got, "ready-f"); s.State != "ready" || len(s.BlockedBy) != 0 {
		t.Fatalf("ready-f (all deps done): got %+v", s)
	}
	blk := stateOf(got, "blk-f")
	if blk.State != "blocked" {
		t.Fatalf("blk-f should be blocked, got %+v", blk)
	}
	joined := strings.Join(blk.BlockedBy, ",")
	if !strings.Contains(joined, "wip-f") || !strings.Contains(joined, "planned-x") {
		t.Fatalf("blk-f BlockedBy must list both unmet deps, got %v", blk.BlockedBy)
	}
}

// Diamond: D ready only when both B and C are done; blocked (by C) otherwise.
func TestDeriveReadiness_Diamond(t *testing.T) {
	deps := map[string][]string{"b": {"a"}, "c": {"a"}, "d": {"b", "c"}}
	rpChdir(t)
	rpSave(t, "a", "done")
	rpSave(t, "b", "done")
	rpSave(t, "c", "done")
	if s := stateOf(roadmap.DeriveReadiness(rpRoadmap(deps, "a", "b", "c", "d")), "d"); s.State != "ready" {
		t.Fatalf("diamond all-done: d should be ready, got %+v", s)
	}
	rpChdir(t)
	rpSave(t, "a", "done")
	rpSave(t, "b", "done")
	d := stateOf(roadmap.DeriveReadiness(rpRoadmap(deps, "a", "b", "c", "d")), "d")
	if d.State != "blocked" || strings.Join(d.BlockedBy, ",") != "c" {
		t.Fatalf("diamond partial: d should be blocked-by c, got %+v", d)
	}
}

// ReadySet returns ready names in declared order; UnmetDependencies lists unmet deps.
func TestReadySetAndUnmetDependencies(t *testing.T) {
	rpChdir(t)
	rpSave(t, "a", "done")
	r := rpRoadmap(map[string][]string{"b": {"a"}, "c": {"a", "z"}}, "a", "b", "c", "z")
	if ready := roadmap.ReadySet(r); strings.Join(ready, ",") != "b,z" {
		t.Fatalf("ReadySet declared order want [b z], got %v", ready)
	}
	if um := roadmap.UnmetDependencies(r, "c"); strings.Join(um, ",") != "z" {
		t.Fatalf("UnmetDependencies(c) want [z], got %v", um)
	}
	if um := roadmap.UnmetDependencies(r, "b"); len(um) != 0 {
		t.Fatalf("UnmetDependencies(b) want none, got %v", um)
	}
	if um := roadmap.UnmetDependencies(nil, "b"); um != nil {
		t.Fatalf("UnmetDependencies(nil) want nil, got %v", um)
	}
	if roadmap.DeriveReadiness(nil) != nil {
		t.Fatal("DeriveReadiness(nil) must be nil")
	}
}

// readinessMarker is exercised through RenderRoadmap (its only call site): ready
// shows 🔓 (ready), blocked shows 🔒 and the blocking dep name, done/in-progress
// carry their annotations and no ready/blocked icon.
func TestRenderRoadmap_MarkersPerState(t *testing.T) {
	rpChdir(t)
	rpSave(t, "done-f", "done")
	rpSave(t, "wip-f", "tests")
	r := rpRoadmap(map[string][]string{"blk-f": {"dep-x"}}, "done-f", "wip-f", "ready-f", "blk-f", "dep-x")
	out := ui.RenderRoadmap(r)
	for _, want := range []string{
		"🔓", "ready-f", "(ready)",
		"🔒", "blk-f", "(blocked-by: dep-x)",
		"✓", "done-f", "(done)",
		"▶", "wip-f", "(in-progress)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderRoadmap missing %q in:\n%s", want, out)
		}
	}
	doneLine := lineWith(out, "done-f")
	if strings.Contains(doneLine, "🔓") || strings.Contains(doneLine, "🔒") {
		t.Fatalf("done feature line must carry no ready/blocked marker: %q", doneLine)
	}
}

// RenderReadyList: populated list prints 🔓 + each name; empty list prints a
// non-empty muted empty-state without any feature names.
func TestRenderReadyList_PopulatedAndEmpty(t *testing.T) {
	out := ui.RenderReadyList([]string{"feature-b", "feature-c"})
	if !strings.Contains(out, "🔓") || !strings.Contains(out, "feature-b") || !strings.Contains(out, "feature-c") {
		t.Fatalf("ready list should show 🔓 + both names, got:\n%s", out)
	}
	empty := ui.RenderReadyList(nil)
	if strings.TrimSpace(empty) == "" {
		t.Fatal("empty ready list must render a non-empty empty-state line")
	}
	if strings.Contains(empty, "feature-b") || strings.Contains(empty, "🔓") {
		t.Fatalf("empty state must not list features or a ready icon, got:\n%s", empty)
	}
}

func lineWith(s, sub string) string {
	for _, ln := range strings.Split(s, "\n") {
		if strings.Contains(ln, sub) {
			return ln
		}
	}
	return ""
}
