// Package roadmapcheckpoint owns the emit/suppress/stale decision for the
// post-roadmap-definition checkpoint directive. All logic here is pure: the
// caller (cmd/centinela) provides a filesystem adapter and the first
// incomplete bootstrap feature; this package never touches os.Stat directly.
package roadmapcheckpoint

import "time"

// Decision enumerates the three outcomes of the checkpoint evaluation.
type Decision int

const (
	// DecisionSuppressed means no directive should be emitted.
	DecisionSuppressed Decision = iota
	// DecisionEmit means the directive should fire for the first time.
	DecisionEmit
	// DecisionStale means the directive should re-fire because the marker
	// is older than at least one roadmap-defining artifact, or is malformed.
	DecisionStale
)

// MarkerPath is the canonical location of the iterate marker.
const MarkerPath = ".workflow/roadmap-checkpoint.json"

// Marker is the on-disk payload that records the user's choice to keep
// iterating on the roadmap definition.
type Marker struct {
	Choice string `json:"choice"`
	At     string `json:"at"`
}

// FS is the narrow filesystem contract Decide needs. It exists so unit tests
// can drive Decide without touching real disk state.
type FS interface {
	// Stat returns the modification time for path. ok=false if the path
	// does not exist; any other error is treated by callers as missing.
	Stat(path string) (time.Time, bool)
	// ReadFile returns the bytes at path. ok=false if missing or unreadable.
	ReadFile(path string) ([]byte, bool)
	// Exists reports whether the given path exists on disk.
	Exists(path string) bool
}

// Decide returns the checkpoint outcome given the current time, the first
// incomplete bootstrap feature (empty + hasFirst=false means suppression),
// and a filesystem adapter. Decide performs no I/O of its own.
func Decide(now time.Time, firstFeature string, hasFirst bool, fs FS) Decision {
	_ = now
	if !hasFirst || firstFeature == "" {
		return DecisionSuppressed
	}
	if fs.Exists(".workflow/" + firstFeature + ".json") {
		return DecisionSuppressed
	}
	data, ok := fs.ReadFile(MarkerPath)
	if !ok {
		return DecisionEmit
	}
	markerAt, parsed := parseMarkerAt(data)
	if !parsed {
		return DecisionStale
	}
	latest, found := LatestMtime(RequiredArtifacts(), fs)
	if !found {
		return DecisionSuppressed
	}
	if latest.After(markerAt) {
		return DecisionStale
	}
	return DecisionSuppressed
}
