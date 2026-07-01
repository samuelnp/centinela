package roadmap

import "testing"

// TestClassifyFeature_Draft returns State "draft" and excludes it from ReadySet.
func TestClassifyFeature_Draft(t *testing.T) {
	if fr := classifyFeature(Feature{Name: "d", Draft: true}); fr.State != "draft" {
		t.Fatalf("draft must classify as draft, got %q", fr.State)
	}
	r := &Roadmap{Phases: []Phase{{Name: "Phase 1", Features: []Feature{
		{Name: "d", Draft: true}, {Name: "r"},
	}}}}
	ready := ReadySet(r)
	if len(ready) != 1 || ready[0] != "r" {
		t.Fatalf("ReadySet must exclude draft, got %v", ready)
	}
}

// TestSummary_ExcludesDraft does not tally an unscored draft as committed work.
func TestSummary_ExcludesDraft(t *testing.T) {
	r := &Roadmap{Phases: []Phase{{Name: "Phase 1", Features: []Feature{
		{Name: "planned-a"}, {Name: "d", Draft: true},
	}}}}
	planned, inProgress, done := r.Summary()
	if planned != 1 || inProgress != 0 || done != 0 {
		t.Fatalf("draft must not be counted: planned=%d ip=%d done=%d", planned, inProgress, done)
	}
}
