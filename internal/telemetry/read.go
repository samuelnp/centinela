package telemetry

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// ReadDefault reads the telemetry log from the default directory.
func ReadDefault() ([]Event, error) { return Read(telemetryDir) }

// Read parses all events from <dir>/events.jsonl. Missing file ⇒ (nil, nil).
// Lenient: lines that fail to unmarshal are skipped (robust to merge artifacts);
// only the valid events are returned.
func Read(dir string) ([]Event, error) {
	f, err := os.Open(filepath.Join(dir, "events.jsonl"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	var events []Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		events = append(events, e)
	}
	if err := sc.Err(); err != nil {
		return events, err
	}
	return events, nil
}
