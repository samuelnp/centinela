package main

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// seedRoadmapRepo chdirs into an isolated temp repo carrying a roadmap, so the
// digest state written by these tests never touches the real .workflow/.
func seedRoadmapRepo(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{
		{Name: "Phase 1", Features: []roadmap.Feature{{Name: "alpha"}, {Name: "beta"}}},
	}}
	if err := roadmap.Save(r); err != nil {
		t.Fatal(err)
	}
}

func hasRoadmapLine(out string) bool { return strings.Contains(out, "Roadmap:") }

func TestEmitRoadmapSummarySuppressesUnchangedRepeat(t *testing.T) {
	seedRoadmapRepo(t)
	first := captureStdout(t, func() { emitRoadmapSummary("s-1") })
	if !hasRoadmapLine(first) {
		t.Fatalf("first call must render, got %q", first)
	}
	second := captureStdout(t, func() { emitRoadmapSummary("s-1") })
	if hasRoadmapLine(second) {
		t.Fatalf("unchanged repeat must be silent, got %q", second)
	}
	// A different session re-renders even though the roadmap is unchanged.
	third := captureStdout(t, func() { emitRoadmapSummary("s-2") })
	if !hasRoadmapLine(third) {
		t.Fatalf("new session must render, got %q", third)
	}
}

func TestEmitRoadmapSummaryRendersAfterRoadmapChange(t *testing.T) {
	seedRoadmapRepo(t)
	captureStdout(t, func() { emitRoadmapSummary("s-1") })
	r, err := roadmap.Load()
	if err != nil {
		t.Fatal(err)
	}
	r.Phases[0].Features = append(r.Phases[0].Features, roadmap.Feature{Name: "gamma"})
	if err := roadmap.Save(r); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() { emitRoadmapSummary("s-1") })
	if !hasRoadmapLine(out) {
		t.Fatalf("changed roadmap must re-render, got %q", out)
	}
}

// AC17: no session signal fails open on every call.
func TestEmitRoadmapSummaryFailsOpenWithoutSession(t *testing.T) {
	seedRoadmapRepo(t)
	for i := 0; i < 2; i++ {
		if out := captureStdout(t, func() { emitRoadmapSummary("") }); !hasRoadmapLine(out) {
			t.Fatalf("call %d without a session id must render, got %q", i, out)
		}
	}
}

// E23: an absent/invalid roadmap prints nothing and writes no digest state.
func TestEmitRoadmapSummaryNoRoadmapWritesNoState(t *testing.T) {
	t.Chdir(t.TempDir())
	if out := captureStdout(t, func() { emitRoadmapSummary("s-1") }); out != "" {
		t.Fatalf("expected no output without a roadmap, got %q", out)
	}
	if _, err := os.Stat(roadmap.SummaryStatePath()); !os.IsNotExist(err) {
		t.Fatalf("digest state must not be written, stat err = %v", err)
	}
}
