package roadmap

import "testing"

// draftReaderRoadmap has a non-draft "a", a draft "d" (both schedulable), and a
// Backlog finding "bk".
func draftReaderRoadmap() *Roadmap {
	return &Roadmap{Phases: []Phase{
		{Name: "Phase 1", Features: []Feature{{Name: "a"}, {Name: "d", Draft: true}}},
		{Name: "Backlog", Features: []Feature{{Name: "bk"}}},
	}}
}

// TestIsDraftFeature reports the persisted draft flag, nil-safe.
func TestIsDraftFeature(t *testing.T) {
	r := draftReaderRoadmap()
	if !IsDraftFeature(r, "d") {
		t.Fatal("d must be a draft")
	}
	if IsDraftFeature(r, "a") {
		t.Fatal("a is not a draft")
	}
	if IsDraftFeature(r, "missing") || IsDraftFeature(nil, "d") {
		t.Fatal("missing feature and nil roadmap must be false")
	}
}

// TestDraftFeatures returns every draft in declared order.
func TestDraftFeatures(t *testing.T) {
	got := DraftFeatures(draftReaderRoadmap())
	if len(got) != 1 || got[0].Name != "d" {
		t.Fatalf("want [d], got %+v", got)
	}
	if DraftFeatures(nil) != nil {
		t.Fatal("nil roadmap → nil drafts")
	}
}

// TestCoverageVsDependencySets contrasts the two schedulable sets: the coverage
// set omits drafts; the dependency-target set includes them.
func TestCoverageVsDependencySets(t *testing.T) {
	r := draftReaderRoadmap()
	cover := NonBacklogFeatureSet(r)
	if !cover["a"] || cover["d"] || cover["bk"] {
		t.Fatalf("coverage set must be {a}: %v", cover)
	}
	deps := dependencyTargetSet(r)
	if !deps["a"] || !deps["d"] || deps["bk"] {
		t.Fatalf("dependency set must be {a,d}: %v", deps)
	}
}
