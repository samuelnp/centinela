package roadmap

import "testing"

func digestFixture() *Roadmap {
	return &Roadmap{Phases: []Phase{
		{Name: "Phase 1", Features: []Feature{{Name: "alpha"}, {Name: "beta"}}},
		{Name: "Phase 2", Features: []Feature{{Name: "gamma"}}},
	}}
}

func TestSummaryDigestIsStableAndNonEmpty(t *testing.T) {
	t.Chdir(t.TempDir())
	d1 := SummaryDigest(digestFixture())
	d2 := SummaryDigest(digestFixture())
	if d1 == "" || d1 != d2 {
		t.Fatalf("digest unstable or empty: %q vs %q", d1, d2)
	}
	if len(d1) != 16 {
		t.Fatalf("expected 16 hex chars, got %d (%q)", len(d1), d1)
	}
}

// The digest hashes the projection, not the bytes: fields the summary never
// shows (intro/note/description prose) must not move it.
func TestSummaryDigestIgnoresNonSummaryProse(t *testing.T) {
	t.Chdir(t.TempDir())
	base := digestFixture()
	reformatted := digestFixture()
	reformatted.Intro = "> some new intro paragraph"
	reformatted.Phases[0].Note = "> a rationale blockquote"
	reformatted.Phases[0].Features[0].Description = "prose churn"
	if SummaryDigest(base) != SummaryDigest(reformatted) {
		t.Fatal("prose-only churn changed the digest")
	}
}

func TestSummaryDigestChangesOnMembershipAndOrder(t *testing.T) {
	t.Chdir(t.TempDir())
	base := SummaryDigest(digestFixture())
	added := digestFixture()
	added.Phases[1].Features = append(added.Phases[1].Features, Feature{Name: "delta"})
	if SummaryDigest(added) == base {
		t.Fatal("adding a feature did not change the digest")
	}
	renamed := digestFixture()
	renamed.Phases[0].Name = "Phase One"
	if SummaryDigest(renamed) == base {
		t.Fatal("renaming a phase did not change the digest")
	}
	reordered := digestFixture()
	reordered.Phases[0].Features[0], reordered.Phases[0].Features[1] =
		reordered.Phases[0].Features[1], reordered.Phases[0].Features[0]
	if SummaryDigest(reordered) == base {
		t.Fatal("declaration order is not part of the digest")
	}
}

func TestSummaryDigestNilRoadmap(t *testing.T) {
	if got := SummaryDigest(nil); got != "" {
		t.Fatalf("nil roadmap should digest to empty, got %q", got)
	}
}

func TestShouldRenderSummaryTruthTable(t *testing.T) {
	prev := SummaryState{SessionID: "s-1", Digest: "d-1"}
	cases := []struct {
		name              string
		sessionID, digest string
		want              bool
	}{
		{"same session, same digest", "s-1", "d-1", false},   // E16
		{"same session, changed digest", "s-1", "d-2", true}, // E17
		{"new session, same digest", "s-2", "d-1", true},     // E18
		{"no session signal fails open", "", "d-1", true},    // E21/AC17
	}
	for _, tc := range cases {
		if got := ShouldRenderSummary(prev, tc.sessionID, tc.digest); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
	// E15: absent state (zero value) always renders.
	if !ShouldRenderSummary(SummaryState{}, "s-1", "d-1") {
		t.Error("absent state must render")
	}
}

func TestSummaryStatePath(t *testing.T) {
	if SummaryStatePath() != ".workflow/.roadmap-digest" {
		t.Fatalf("unexpected state path %q", SummaryStatePath())
	}
}
