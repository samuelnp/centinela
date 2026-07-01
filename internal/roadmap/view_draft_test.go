package roadmap

import "testing"

// TestBuildView_Draft flows a draft through as draft:true + readiness:"draft"
// and excludes it from the committed counts, while a non-draft is unaffected.
func TestBuildView_Draft(t *testing.T) {
	r := &Roadmap{Phases: []Phase{{Name: "Phase 1", Features: []Feature{
		{Name: "d", Draft: true}, {Name: "planned-a"},
	}}}}
	v := BuildView(r)
	feats := v.Phases[0].Features
	var draft, plain FeatureView
	for _, f := range feats {
		if f.Name == "d" {
			draft = f
		} else {
			plain = f
		}
	}
	if !draft.Draft || draft.Readiness != "draft" {
		t.Fatalf("draft view must set draft/readiness: %+v", draft)
	}
	if plain.Draft || plain.Readiness == "draft" {
		t.Fatalf("non-draft must not carry draft signals: %+v", plain)
	}
	if v.Counts.Planned != 1 {
		t.Fatalf("draft must be excluded from counts, planned=%d", v.Counts.Planned)
	}
}
