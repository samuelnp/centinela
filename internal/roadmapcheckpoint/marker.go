package roadmapcheckpoint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ReadMarker reads and parses the iterate marker at path. A missing file
// returns (nil, nil) so callers can distinguish "absent" from "broken".
// A malformed JSON or unparseable timestamp returns an error and leaves
// the decision logic to treat the marker as stale.
func ReadMarker(path string) (*Marker, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var m Marker
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if _, err := time.Parse(time.RFC3339, m.At); err != nil {
		return &m, err
	}
	return &m, nil
}

// WriteMarker persists an iterate marker at path with the supplied timestamp.
// The parent directory is created if missing. The timestamp is serialized
// in RFC 3339 so freshness comparisons are stable across timezones.
func WriteMarker(path string, now time.Time) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	m := Marker{Choice: "iterate", At: now.UTC().Format(time.RFC3339)}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// parseMarkerAt extracts the marker's RFC 3339 timestamp from raw bytes.
// Returns ok=false when the JSON is malformed or the timestamp is missing
// or unparseable, signaling the caller to treat the marker as stale.
func parseMarkerAt(data []byte) (time.Time, bool) {
	var m Marker
	if err := json.Unmarshal(data, &m); err != nil {
		return time.Time{}, false
	}
	if m.At == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, m.At)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
