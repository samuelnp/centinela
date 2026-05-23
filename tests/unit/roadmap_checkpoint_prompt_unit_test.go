package unit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/roadmap"
	rc "github.com/samuelnp/centinela/internal/roadmapcheckpoint"
)

// These tests validate the roadmapcheckpoint PUBLIC contract from a consumer's
// perspective (the cmd/centinela hook drives it the same way). The exhaustive
// per-branch coverage lives in the in-package test
// internal/roadmapcheckpoint/checkpoint_decide_test.go; this file pins the
// observable Decide outcomes a caller depends on through a fake FS.

// fakeFS is an in-memory FS so Decide can be driven without touching disk.
type fakeFS struct {
	mtimes map[string]time.Time
	bytes  map[string][]byte
}

func newFakeFS() *fakeFS {
	return &fakeFS{mtimes: map[string]time.Time{}, bytes: map[string][]byte{}}
}

func (f *fakeFS) Stat(p string) (time.Time, bool)  { mt, ok := f.mtimes[p]; return mt, ok }
func (f *fakeFS) ReadFile(p string) ([]byte, bool) { b, ok := f.bytes[p]; return b, ok }
func (f *fakeFS) Exists(p string) bool {
	if _, ok := f.bytes[p]; ok {
		return true
	}
	_, ok := f.mtimes[p]
	return ok
}

func (f *fakeFS) stampArtifacts(mt time.Time) {
	for _, p := range rc.RequiredArtifacts() {
		f.mtimes[p] = mt
	}
}

const firstFeature = "phase-0-feature-a"

var baseTime = time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)

func markerJSON(at string) []byte {
	return []byte(`{"choice":"iterate","at":"` + at + `"}`)
}

// TestDecide_PublicOutcomes pins each observable Decision a caller relies on.
func TestDecide_PublicOutcomes(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*fakeFS)
		want  rc.Decision
	}{
		{"no marker -> emit", func(f *fakeFS) { f.stampArtifacts(baseTime) }, rc.DecisionEmit},
		{"marker at == latest -> suppress", func(f *fakeFS) {
			f.stampArtifacts(baseTime)
			f.bytes[rc.MarkerPath] = markerJSON(baseTime.Format(time.RFC3339))
		}, rc.DecisionSuppressed},
		{"marker after artifacts -> suppress", func(f *fakeFS) {
			f.stampArtifacts(baseTime)
			f.bytes[rc.MarkerPath] = markerJSON(baseTime.Add(time.Hour).Format(time.RFC3339))
		}, rc.DecisionSuppressed},
		{"ROADMAP newer -> stale", func(f *fakeFS) {
			f.stampArtifacts(baseTime)
			f.mtimes["ROADMAP.md"] = baseTime.Add(time.Hour)
			f.bytes[rc.MarkerPath] = markerJSON(baseTime.Format(time.RFC3339))
		}, rc.DecisionStale},
		{"analysis artifact newer -> stale", func(f *fakeFS) {
			f.stampArtifacts(baseTime)
			f.mtimes[".workflow/roadmap-analysis.json"] = baseTime.Add(time.Hour)
			f.bytes[rc.MarkerPath] = markerJSON(baseTime.Format(time.RFC3339))
		}, rc.DecisionStale},
		{"malformed marker JSON -> stale", func(f *fakeFS) {
			f.stampArtifacts(baseTime)
			f.bytes[rc.MarkerPath] = []byte("{not valid json")
		}, rc.DecisionStale},
		{"unparseable at -> stale", func(f *fakeFS) {
			f.stampArtifacts(baseTime)
			f.bytes[rc.MarkerPath] = markerJSON("not-a-timestamp")
		}, rc.DecisionStale},
		{"workflow file exists -> suppress", func(f *fakeFS) {
			f.stampArtifacts(baseTime)
			f.bytes[".workflow/"+firstFeature+".json"] = []byte("{}")
		}, rc.DecisionSuppressed},
		{"valid marker but no artifacts -> suppress", func(f *fakeFS) {
			f.bytes[rc.MarkerPath] = markerJSON(baseTime.Format(time.RFC3339))
		}, rc.DecisionSuppressed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFakeFS()
			c.setup(f)
			if got := rc.Decide(baseTime, firstFeature, true, f); got != c.want {
				t.Fatalf("Decide = %v, want %v", got, c.want)
			}
		})
	}
}

func TestDecide_NoFirstFeature_Suppressed(t *testing.T) {
	f := newFakeFS()
	f.stampArtifacts(baseTime)
	if got := rc.Decide(baseTime, "", false, f); got != rc.DecisionSuppressed {
		t.Fatalf("hasFirst=false should Suppress, got %v", got)
	}
	if got := rc.Decide(baseTime, "", true, f); got != rc.DecisionSuppressed {
		t.Fatalf("empty firstFeature should Suppress, got %v", got)
	}
}

// --- FirstIncompleteBootstrap public behavior (status read from cwd) ---

func bootstrapRoadmap(features ...string) *roadmap.Roadmap {
	var fs []roadmap.Feature
	for _, n := range features {
		fs = append(fs, roadmap.Feature{Name: n})
	}
	return &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 0: Bootstrap", Features: fs}}}
}

func chdirTemp(t *testing.T) {
	t.Helper()
	d := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeWorkflowStep(t *testing.T, feature, step string) {
	t.Helper()
	body := `{"feature":"` + feature + `","currentStep":"` + step + `","steps":{}}`
	if err := os.WriteFile(filepath.Join(".workflow", feature+".json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFirstIncompleteBootstrap_PublicBehavior(t *testing.T) {
	if name, ok := rc.FirstIncompleteBootstrap(nil); ok || name != "" {
		t.Fatalf("nil roadmap -> (\"\", false), got (%q, %v)", name, ok)
	}

	chdirTemp(t)

	// First non-done picked.
	if name, ok := rc.FirstIncompleteBootstrap(bootstrapRoadmap("alpha", "beta")); !ok || name != "alpha" {
		t.Fatalf("expected (alpha, true), got (%q, %v)", name, ok)
	}

	// Done feature skipped, second picked.
	writeWorkflowStep(t, "alpha", "done")
	if name, ok := rc.FirstIncompleteBootstrap(bootstrapRoadmap("alpha", "beta")); !ok || name != "beta" {
		t.Fatalf("expected (beta, true) when alpha done, got (%q, %v)", name, ok)
	}

	// All done -> none.
	writeWorkflowStep(t, "beta", "done")
	if name, ok := rc.FirstIncompleteBootstrap(bootstrapRoadmap("alpha", "beta")); ok || name != "" {
		t.Fatalf("all done -> (\"\", false), got (%q, %v)", name, ok)
	}
}

func TestWriteThenReadMarker_RoundTrip(t *testing.T) {
	chdirTemp(t)
	if err := rc.WriteMarker(rc.MarkerPath, baseTime); err != nil {
		t.Fatalf("WriteMarker: %v", err)
	}
	m, err := rc.ReadMarker(rc.MarkerPath)
	if err != nil || m == nil || m.Choice != "iterate" {
		t.Fatalf("expected choice=iterate, got (%+v, %v)", m, err)
	}
	if _, err := time.Parse(time.RFC3339, m.At); err != nil {
		t.Fatalf("marker At not RFC3339: %q (%v)", m.At, err)
	}
}
