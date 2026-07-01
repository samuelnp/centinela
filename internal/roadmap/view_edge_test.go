package roadmap

import (
	"encoding/json"
	"testing"
)

// BuildView marshals byte-identically across two builds of the same roadmap.
func TestBuildView_ByteStable(t *testing.T) {
	chdirRoadmapTemp(t)
	seedDone(t, "auth-service")
	seedStep(t, "billing-api", "code")
	a, err := json.MarshalIndent(BuildView(canonicalView()), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(BuildView(canonicalView()), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Fatalf("byte-unstable:\n%s\n---\n%s", a, b)
	}
}

// Empty and nil roadmaps yield empty phases with all-zero counts, marshaling to
// the exact contract bytes {"phases":[],"counts":{...0...}}.
func TestBuildView_EmptyAndNil(t *testing.T) {
	chdirRoadmapTemp(t)
	const want = `{"phases":[],"counts":{"planned":0,"inProgress":0,"done":0}}`
	for _, r := range []*Roadmap{{}, nil} {
		v := BuildView(r)
		if len(v.Phases) != 0 || v.Counts != (StatusCounts{}) {
			t.Fatalf("empty roadmap must yield empty phases + zero counts, got %+v", v)
		}
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != want {
			t.Fatalf("empty view JSON = %s want %s", data, want)
		}
	}
}

// Backlog/Baseline phases are excluded from both the phases list and the counts.
func TestBuildView_ExcludesNonSchedulable(t *testing.T) {
	chdirRoadmapTemp(t)
	r := &Roadmap{Phases: []Phase{
		{Name: "Backlog", Features: []Feature{{Name: "bl"}}},
		{Name: "Baseline", Features: []Feature{{Name: "bs"}}},
		{Name: "Q1", Features: []Feature{{Name: "q"}}},
	}}
	v := BuildView(r)
	if len(v.Phases) != 1 || v.Phases[0].Name != "Q1" {
		t.Fatalf("non-schedulable phases must be excluded: %+v", v.Phases)
	}
	if v.Counts != (StatusCounts{Planned: 1}) {
		t.Fatalf("counts must reflect only schedulable Q1: %+v", v.Counts)
	}
}
