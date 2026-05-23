package roadmapcheckpoint

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// memFS is an in-memory FS so Decide/LatestMtime can be driven deterministically.
type memFS struct {
	mtimes map[string]time.Time
	bytes  map[string][]byte
}

func newMemFS() *memFS {
	return &memFS{mtimes: map[string]time.Time{}, bytes: map[string][]byte{}}
}

func (m *memFS) Stat(p string) (time.Time, bool)  { mt, ok := m.mtimes[p]; return mt, ok }
func (m *memFS) ReadFile(p string) ([]byte, bool) { b, ok := m.bytes[p]; return b, ok }
func (m *memFS) Exists(p string) bool {
	if _, ok := m.bytes[p]; ok {
		return true
	}
	_, ok := m.mtimes[p]
	return ok
}

func (m *memFS) stampArtifacts(mt time.Time) {
	for _, p := range RequiredArtifacts() {
		m.mtimes[p] = mt
	}
}

const feat = "phase-0-feature-a"

var t0 = time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)

func marker(at string) []byte { return []byte(`{"choice":"iterate","at":"` + at + `"}`) }

func TestDecide_AllOutcomes(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*memFS)
		want  Decision
	}{
		{"no marker emits", func(m *memFS) { m.stampArtifacts(t0) }, DecisionEmit},
		{"fresh equal suppresses", func(m *memFS) {
			m.stampArtifacts(t0)
			m.bytes[MarkerPath] = marker(t0.Format(time.RFC3339))
		}, DecisionSuppressed},
		{"fresh after suppresses", func(m *memFS) {
			m.stampArtifacts(t0)
			m.bytes[MarkerPath] = marker(t0.Add(time.Hour).Format(time.RFC3339))
		}, DecisionSuppressed},
		{"roadmap newer stale", func(m *memFS) {
			m.stampArtifacts(t0)
			m.mtimes["ROADMAP.md"] = t0.Add(time.Hour)
			m.bytes[MarkerPath] = marker(t0.Format(time.RFC3339))
		}, DecisionStale},
		{"malformed json stale", func(m *memFS) {
			m.stampArtifacts(t0)
			m.bytes[MarkerPath] = []byte("{nope")
		}, DecisionStale},
		{"unparseable at stale", func(m *memFS) {
			m.stampArtifacts(t0)
			m.bytes[MarkerPath] = marker("nope")
		}, DecisionStale},
		{"workflow file suppresses", func(m *memFS) {
			m.stampArtifacts(t0)
			m.bytes[".workflow/"+feat+".json"] = []byte("{}")
		}, DecisionSuppressed},
		{"no artifacts suppresses", func(m *memFS) {
			m.bytes[MarkerPath] = marker(t0.Format(time.RFC3339))
		}, DecisionSuppressed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := newMemFS()
			c.setup(m)
			if got := Decide(t0, feat, true, m); got != c.want {
				t.Fatalf("Decide = %v, want %v", got, c.want)
			}
		})
	}
}

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

// --- FirstIncompleteBootstrap (reads .workflow/<name>.json from cwd) ---

func chdirTmp(t *testing.T) {
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

func bootstrap(features ...string) *roadmap.Roadmap {
	var fs []roadmap.Feature
	for _, n := range features {
		fs = append(fs, roadmap.Feature{Name: n})
	}
	return &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 0: Bootstrap", Features: fs}}}
}

func setStep(t *testing.T, feature, step string) {
	t.Helper()
	body := `{"feature":"` + feature + `","currentStep":"` + step + `","steps":{}}`
	if err := os.WriteFile(filepath.Join(".workflow", feature+".json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFirstIncompleteBootstrap_Cases(t *testing.T) {
	if name, ok := FirstIncompleteBootstrap(nil); ok || name != "" {
		t.Fatalf("nil roadmap -> (\"\", false), got (%q, %v)", name, ok)
	}

	chdirTmp(t)

	// No bootstrap phase.
	noBoot := &roadmap.Roadmap{Phases: []roadmap.Phase{{Name: "Phase 1", Features: []roadmap.Feature{{Name: "x"}}}}}
	if name, ok := FirstIncompleteBootstrap(noBoot); ok || name != "" {
		t.Fatalf("no bootstrap -> (\"\", false), got (%q, %v)", name, ok)
	}

	// First non-done picked.
	if name, ok := FirstIncompleteBootstrap(bootstrap("alpha", "beta")); !ok || name != "alpha" {
		t.Fatalf("expected alpha, got (%q, %v)", name, ok)
	}

	// alpha done -> beta.
	setStep(t, "alpha", "done")
	if name, ok := FirstIncompleteBootstrap(bootstrap("alpha", "beta")); !ok || name != "beta" {
		t.Fatalf("expected beta, got (%q, %v)", name, ok)
	}

	// in-progress is non-done.
	setStep(t, "beta", "code")
	if name, ok := FirstIncompleteBootstrap(bootstrap("beta")); !ok || name != "beta" {
		t.Fatalf("in-progress beta is incomplete, got (%q, %v)", name, ok)
	}

	// all done -> none.
	setStep(t, "beta", "done")
	if name, ok := FirstIncompleteBootstrap(bootstrap("alpha", "beta")); ok || name != "" {
		t.Fatalf("all done -> (\"\", false), got (%q, %v)", name, ok)
	}
}

// --- WriteMarker / ReadMarker / NewOSFS over real disk ---

func TestMarkerRoundTripAndOSFS(t *testing.T) {
	chdirTmp(t)

	if m, err := ReadMarker(MarkerPath); m != nil || err != nil {
		t.Fatalf("missing marker -> (nil,nil), got (%+v,%v)", m, err)
	}

	if err := WriteMarker(MarkerPath, t0); err != nil {
		t.Fatalf("WriteMarker: %v", err)
	}
	m, err := ReadMarker(MarkerPath)
	if err != nil || m == nil || m.Choice != "iterate" {
		t.Fatalf("ReadMarker -> (%+v, %v)", m, err)
	}
	if _, err := time.Parse(time.RFC3339, m.At); err != nil {
		t.Fatalf("marker At not RFC3339: %q", m.At)
	}

	// Malformed marker on disk -> ReadMarker returns an error.
	if err := os.WriteFile(MarkerPath, []byte("{bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadMarker(MarkerPath); err == nil {
		t.Fatal("malformed marker should error from ReadMarker")
	}

	// Unparseable at -> ReadMarker returns marker AND error.
	if err := os.WriteFile(MarkerPath, marker("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if mm, err := ReadMarker(MarkerPath); err == nil || mm == nil {
		t.Fatalf("unparseable at should return (marker, err), got (%+v, %v)", mm, err)
	}

	// NewOSFS exercises Stat/ReadFile/Exists against real disk.
	fs := NewOSFS()
	if !fs.Exists(MarkerPath) {
		t.Fatal("OSFS.Exists should see the marker")
	}
	if _, ok := fs.Stat(MarkerPath); !ok {
		t.Fatal("OSFS.Stat should see the marker")
	}
	if _, ok := fs.ReadFile(MarkerPath); !ok {
		t.Fatal("OSFS.ReadFile should read the marker")
	}
	if fs.Exists("nonexistent-path-xyz") {
		t.Fatal("OSFS.Exists should be false for missing path")
	}
	if _, ok := fs.Stat("nonexistent-path-xyz"); ok {
		t.Fatal("OSFS.Stat should be false for missing path")
	}
	if _, ok := fs.ReadFile("nonexistent-path-xyz"); ok {
		t.Fatal("OSFS.ReadFile should be false for missing path")
	}
}
