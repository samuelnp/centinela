package roadmapstate

import "testing"

func TestPathsIsTheAlwaysPresentPair(t *testing.T) {
	got := Paths()
	want := []string{".workflow/roadmap.json", "ROADMAP.md"}
	if len(got) != len(want) {
		t.Fatalf("Paths() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Paths()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// Mutating the returned slice must not poison the next caller's pathspec.
func TestPathsReturnsAFreshSlice(t *testing.T) {
	first := Paths()
	first[0] = "hacked"
	if Paths()[0] != RoadmapJSON {
		t.Fatal("Paths() must not hand out shared backing storage")
	}
}

func TestIsStatePath(t *testing.T) {
	for _, p := range []string{
		".workflow/roadmap.json", "ROADMAP.md", ".workflow/roadmap-analysis.json",
		".workflow/roadmap-quality.md", ".workflow/f-gatekeeper.md",
		` "ROADMAP.md" `, ".workflow/nested/deep.json",
	} {
		if !IsStatePath(p) {
			t.Errorf("IsStatePath(%q) = false, want true", p)
		}
	}
	for _, p := range []string{
		"", "   ", "internal/x.go", "docs/ROADMAP.md", "ROADMAP.md.bak",
		"workflow/roadmap.json", "a/.workflow/roadmap.json",
	} {
		if IsStatePath(p) {
			t.Errorf("IsStatePath(%q) = true, want false", p)
		}
	}
}

// The strict-subset property is the safety guarantee: Covers is never a filter.
func TestCoversIsAStrictSubsetTest(t *testing.T) {
	if !Covers([]string{".workflow/roadmap.json", "ROADMAP.md"}) {
		t.Fatal("all-state set must be covered")
	}
	if Covers([]string{".workflow/roadmap.json", "internal/x.go"}) {
		t.Fatal("a mixed set must NOT be covered — that is the freshness hole")
	}
	if Covers([]string{"internal/x.go"}) {
		t.Fatal("a source-only set must not be covered")
	}
}

// Empty is deliberately not covered; "nothing changed" is a different question.
func TestCoversEmptyIsFalse(t *testing.T) {
	if Covers(nil) || Covers([]string{}) {
		t.Fatal("an empty set must not report as covered")
	}
}
