package roadmap

import "testing"

func TestIsBaselinePhaseName_MatchesAndRejects(t *testing.T) {
	for _, ok := range []string{"Baseline", " baseline ", "BASELINE"} {
		if !isBaselinePhaseName(ok) || !IsBaselinePhaseName(ok) {
			t.Fatalf("%q must be recognized as Baseline", ok)
		}
	}
	for _, no := range []string{"Backlog", "Phase 1", "baselines"} {
		if isBaselinePhaseName(no) {
			t.Fatalf("%q must NOT be Baseline", no)
		}
	}
}

func TestIsNonSchedulablePhase_BacklogAndBaseline(t *testing.T) {
	for _, name := range []string{"Backlog", "Baseline"} {
		if !isNonSchedulablePhase(name) {
			t.Fatalf("%q must be non-schedulable", name)
		}
	}
	if isNonSchedulablePhase("Phase 0: Bootstrap") {
		t.Fatal("regular phases must remain schedulable")
	}
}

// roadmapWithBaseline pairs a Baseline phase with one schedulable phase.
func roadmapWithBaseline() *Roadmap {
	return &Roadmap{Phases: []Phase{
		{Name: BaselinePhaseName, Features: []Feature{{Name: "built-a"}, {Name: "built-b"}}},
		{Name: "Phase 1", Features: []Feature{{Name: "real-work"}}},
	}}
}

func TestSummary_ExcludesBaseline(t *testing.T) {
	planned, inProgress, done := roadmapWithBaseline().Summary()
	if planned != 1 || inProgress != 0 || done != 0 {
		t.Fatalf("Baseline features must be excluded from counts, got %d/%d/%d", planned, inProgress, done)
	}
}

func TestNonBacklogFeatureSet_ExcludesBaseline(t *testing.T) {
	set := NonBacklogFeatureSet(roadmapWithBaseline())
	if set["built-a"] || set["built-b"] {
		t.Fatal("Baseline features must be excluded from the coverage set")
	}
	if !set["real-work"] {
		t.Fatal("schedulable features must remain in the coverage set")
	}
}

func TestDeriveReadiness_ExcludesBaseline(t *testing.T) {
	for _, fr := range DeriveReadiness(roadmapWithBaseline()) {
		if fr.Name == "built-a" || fr.Name == "built-b" {
			t.Fatalf("Baseline feature %q must not appear in readiness", fr.Name)
		}
	}
}

// TestBacklogBehaviorUnchanged guards the existing Backlog exemption regression.
func TestBacklogBehaviorUnchanged(t *testing.T) {
	r := &Roadmap{Phases: []Phase{
		{Name: BacklogPhaseName, Features: []Feature{{Name: "deferred"}}},
		{Name: "Phase 1", Features: []Feature{{Name: "real"}}},
	}}
	planned, _, _ := r.Summary()
	if planned != 1 {
		t.Fatalf("Backlog must stay excluded from counts, got %d", planned)
	}
	if NonBacklogFeatureSet(r)["deferred"] {
		t.Fatal("Backlog must stay excluded from the coverage set")
	}
}
