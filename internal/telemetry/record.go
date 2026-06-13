package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

// telemetryDir/eventsFile own the storage path (hardcoded so telemetry never
// imports internal/workflow and stays a config-only leaf).
const telemetryDir = ".workflow/telemetry"
const eventsFile = ".workflow/telemetry/events.jsonl"

// now is overridable in tests for deterministic timestamps.
var now = func() time.Time { return time.Now().UTC() }

// Record appends one event, best-effort. No-op when telemetry is disabled or
// cfg is nil. Stamps Schema + Timestamp. I/O errors warn to stderr and are
// swallowed — recording NEVER fails the host command (mirrors internal/memory).
func Record(cfg *config.Config, e Event) {
	if cfg == nil || !cfg.Telemetry.IsEnabled() {
		return
	}
	e.Schema = Schema
	if e.Timestamp == "" {
		e.Timestamp = now().Format(time.RFC3339)
	}
	if err := appendEvent(e); err != nil {
		fmt.Fprintln(os.Stderr, "[telemetry] warning: "+err.Error())
	}
}

func appendEvent(e Event) error {
	if err := os.MkdirAll(telemetryDir, 0o755); err != nil {
		return err
	}
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(eventsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck
	_, err = f.Write(append(line, '\n'))
	return err
}
