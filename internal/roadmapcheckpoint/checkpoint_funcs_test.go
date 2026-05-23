package roadmapcheckpoint

import (
	"testing"
	"time"
)

func TestDecide_NoFirstSuppresses(t *testing.T) {
	m := newMemFS()
	m.stampArtifacts(t0)
	if got := Decide(t0, "", false, m); got != DecisionSuppressed {
		t.Fatalf("hasFirst=false should suppress, got %v", got)
	}
	if got := Decide(t0, "", true, m); got != DecisionSuppressed {
		t.Fatalf("empty feature should suppress, got %v", got)
	}
}

func TestLatestMtime_PicksMax(t *testing.T) {
	m := newMemFS()
	m.mtimes["ROADMAP.md"] = t0
	m.mtimes[".workflow/roadmap.json"] = t0.Add(time.Hour)
	latest, found := LatestMtime(RequiredArtifacts(), m)
	if !found || !latest.Equal(t0.Add(time.Hour)) {
		t.Fatalf("expected latest = t0+1h, got (%v, %v)", latest, found)
	}
	if _, found := LatestMtime(RequiredArtifacts(), newMemFS()); found {
		t.Fatal("no artifacts should yield found=false")
	}
}

func TestParseMarkerAt(t *testing.T) {
	if _, ok := parseMarkerAt(marker(t0.Format(time.RFC3339))); !ok {
		t.Fatal("valid marker should parse")
	}
	if _, ok := parseMarkerAt([]byte("{bad")); ok {
		t.Fatal("malformed json should not parse")
	}
	if _, ok := parseMarkerAt([]byte(`{"choice":"iterate"}`)); ok {
		t.Fatal("empty at should not parse")
	}
	if _, ok := parseMarkerAt(marker("nope")); ok {
		t.Fatal("unparseable at should not parse")
	}
}

func TestRequiredArtifacts_Canonical(t *testing.T) {
	got := RequiredArtifacts()
	if len(got) != 6 {
		t.Fatalf("expected 6 required artifacts, got %d: %v", len(got), got)
	}
}
