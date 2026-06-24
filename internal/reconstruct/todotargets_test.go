package reconstruct

import "testing"

func TestTodoTargets_ReturnsTodoBearingInTargetOrder(t *testing.T) {
	r := Reconstruction{
		Targets: []Target{{Slug: "a"}, {Slug: "b"}, {Slug: "c"}},
		Features: []Artifact{
			{Slug: "a", Body: "Feature: a\n" + todoMarker},
			{Slug: "b", Body: "Feature: b\nconfirmed"},
			{Slug: "c", Body: "Feature: c\n" + todoMarker},
		},
	}
	got := r.TodoTargets()
	if len(got) != 2 || got[0].Slug != "a" || got[1].Slug != "c" {
		t.Fatalf("expected the TODO-bearing targets a,c in order, got %+v", got)
	}
}

func TestTodoTargets_ZeroWhenNoMarkers(t *testing.T) {
	r := Reconstruction{
		Targets:  []Target{{Slug: "a"}},
		Features: []Artifact{{Slug: "a", Body: "Feature: a\nall confirmed"}},
	}
	if got := r.TodoTargets(); got != nil {
		t.Fatalf("expected nil when no artifact carries a TODO marker, got %+v", got)
	}
}
