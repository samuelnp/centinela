package unit_test

import (
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

func phase(name string, features ...string) roadmap.Phase {
	p := roadmap.Phase{Name: name}
	for _, f := range features {
		p.Features = append(p.Features, roadmap.Feature{Name: f})
	}
	return p
}

// markDone seeds a done workflow-state file so FeatureStatus(name) == "done".
func markDone(t *testing.T, name string) {
	t.Helper()
	wf := workflow.New(name)
	wf.CurrentStep = "done"
	if err := workflow.Save(wf); err != nil {
		t.Fatalf("save %s: %v", name, err)
	}
}

// Spec Half-B scenario: first incomplete is across ALL phases — when every
// Phase 0 feature is done, the walk continues into Phase 1.
func TestFirstIncomplete_CrossesPhases(t *testing.T) {
	chdirWorkflowTemp(t)
	markDone(t, "p0-a")
	markDone(t, "p0-b")
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{
		phase("Phase 0", "p0-a", "p0-b"),
		phase("Phase 1", "phase-1-first", "phase-1-second"),
	}}
	name, ok := roadmap.FirstIncomplete(r)
	if !ok {
		t.Fatal("expected an incomplete feature, got none")
	}
	if name != "phase-1-first" {
		t.Fatalf("expected phase-1-first, got %q", name)
	}
}

// Spec roadmap-complete scenario: every feature done -> ("", false).
func TestFirstIncomplete_AllDoneReturnsFalse(t *testing.T) {
	chdirWorkflowTemp(t)
	markDone(t, "x")
	markDone(t, "y")
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{phase("P", "x", "y")}}
	name, ok := roadmap.FirstIncomplete(r)
	if ok || name != "" {
		t.Fatalf("all-done should yield (\"\", false), got (%q, %v)", name, ok)
	}
}

// nil and empty roadmaps yield ("", false).
func TestFirstIncomplete_NilAndEmpty(t *testing.T) {
	if name, ok := roadmap.FirstIncomplete(nil); ok || name != "" {
		t.Fatalf("nil roadmap: want (\"\", false), got (%q, %v)", name, ok)
	}
	empty := &roadmap.Roadmap{}
	if name, ok := roadmap.FirstIncomplete(empty); ok || name != "" {
		t.Fatalf("empty roadmap: want (\"\", false), got (%q, %v)", name, ok)
	}
}

// FirstNotDone: a feature with no workflow file is "planned" (not done) -> true;
// a done feature -> ("", false).
func TestFirstNotDone_Predicate(t *testing.T) {
	chdirWorkflowTemp(t)
	if name, ok := roadmap.FirstNotDone("never-started"); !ok || name != "never-started" {
		t.Fatalf("planned feature should be not-done, got (%q, %v)", name, ok)
	}
	markDone(t, "finished")
	if name, ok := roadmap.FirstNotDone("finished"); ok || name != "" {
		t.Fatalf("done feature should yield (\"\", false), got (%q, %v)", name, ok)
	}
}
