package roadmap

import "testing"

// TestIsBacklogPhaseName covers exact match, case variants, trim, and partial.
func TestIsBacklogPhaseName(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"Backlog", true},
		{"backlog", true},
		{"BACKLOG", true},
		{"  Backlog  ", true},
		{"Backlog Phase Items", false},
		{"Pre-Backlog Work", false},
		{"", false},
		{"Phase 0: Bootstrap", false},
	}
	for _, c := range cases {
		got := isBacklogPhaseName(c.name)
		if got != c.want {
			t.Errorf("isBacklogPhaseName(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestIsBacklogPhaseName_Exported delegates to isBacklogPhaseName.
func TestIsBacklogPhaseName_Exported(t *testing.T) {
	if !IsBacklogPhaseName("backlog") {
		t.Error("exported form should accept lowercase backlog")
	}
	if IsBacklogPhaseName("Pre-Backlog Work") {
		t.Error("partial match must not be exempt")
	}
}

// TestNonBacklogFeatureSet_ExcludesBacklog checks that Backlog features
// are absent from the coverage set and real features are present.
func TestNonBacklogFeatureSet_ExcludesBacklog(t *testing.T) {
	r := &Roadmap{Phases: []Phase{
		{Name: "Phase 0", Features: []Feature{{Name: "real"}}},
		{Name: "Backlog", Features: []Feature{{Name: "deferred"}}},
	}}
	set := NonBacklogFeatureSet(r)
	if !set["real"] {
		t.Error("real phase feature must be in coverage set")
	}
	if set["deferred"] {
		t.Error("Backlog feature must be excluded from coverage set")
	}
}

// TestNonBacklogFeatureSet_NilRoadmap returns an empty set.
func TestNonBacklogFeatureSet_NilRoadmap(t *testing.T) {
	if s := NonBacklogFeatureSet(nil); len(s) != 0 {
		t.Errorf("expected empty set for nil roadmap, got %v", s)
	}
}

// TestBacklogFeatures covers normal and nil cases.
func TestBacklogFeatures(t *testing.T) {
	if BacklogFeatures(nil) != nil {
		t.Error("nil roadmap should return nil")
	}
	r := &Roadmap{Phases: []Phase{
		{Name: "Phase 0", Features: []Feature{{Name: "real"}}},
		{Name: "Backlog", Features: []Feature{{Name: "d1"}, {Name: "d2"}}},
	}}
	got := BacklogFeatures(r)
	if len(got) != 2 || got[0].Name != "d1" || got[1].Name != "d2" {
		t.Errorf("unexpected BacklogFeatures: %v", got)
	}
}

// TestIsBacklogFeature covers nil roadmap and membership.
func TestIsBacklogFeature(t *testing.T) {
	if IsBacklogFeature(nil, "x") {
		t.Error("nil roadmap must return false")
	}
	r := &Roadmap{Phases: []Phase{
		{Name: "Backlog", Features: []Feature{{Name: "deferred-slug"}}},
		{Name: "Phase 1", Features: []Feature{{Name: "real-slug"}}},
	}}
	if !IsBacklogFeature(r, "deferred-slug") {
		t.Error("deferred-slug must be a Backlog feature")
	}
	if IsBacklogFeature(r, "real-slug") {
		t.Error("real-slug is not a Backlog feature")
	}
}
