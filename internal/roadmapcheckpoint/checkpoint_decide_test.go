package roadmapcheckpoint

import (
	"testing"
	"time"
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
