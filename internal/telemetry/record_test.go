package telemetry

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
)

// fixedNow overrides the package now var for deterministic timestamps and
// restores it when the test finishes.
func fixedNow(t *testing.T, ts string) {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatalf("bad fixed time %q: %v", ts, err)
	}
	old := now
	now = func() time.Time { return parsed }
	t.Cleanup(func() { now = old })
}

func enabledCfg() *config.Config { return &config.Config{} }

func disabledCfg() *config.Config {
	f := false
	return &config.Config{Telemetry: config.TelemetryConfig{Enabled: &f}}
}

// TestRecord_GoldenLine — enabled Record writes the exact JSON line with the
// schema id and the deterministic RFC3339 timestamp.
func TestRecord_GoldenLine(t *testing.T) {
	t.Chdir(t.TempDir())
	fixedNow(t, "2026-06-13T12:00:00Z")
	Record(enabledCfg(), Event{Type: TypeGateFailure, Gate: "G1", Message: "too big"})

	got, err := os.ReadFile(filepath.Join(telemetryDir, "events.jsonl"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	want := `{"schema":"centinela.telemetry/v1","type":"gate-failure","timestamp":"2026-06-13T12:00:00Z","gate":"G1","message":"too big"}` + "\n"
	if string(got) != want {
		t.Fatalf("golden line mismatch:\n got: %q\nwant: %q", got, want)
	}
}

// TestRecord_DisabledNoOp — disabled telemetry writes no file at all.
func TestRecord_DisabledNoOp(t *testing.T) {
	t.Chdir(t.TempDir())
	Record(disabledCfg(), Event{Type: TypeStepAdvanced, Feature: "alpha", Step: "plan"})
	if _, err := os.Stat(eventsFile); !os.IsNotExist(err) {
		t.Fatalf("expected no telemetry file when disabled, err=%v", err)
	}
}

// TestRecord_NilCfgNoOp — nil config is a no-op (no file).
func TestRecord_NilCfgNoOp(t *testing.T) {
	t.Chdir(t.TempDir())
	Record(nil, Event{Type: TypeBlock})
	if _, err := os.Stat(eventsFile); !os.IsNotExist(err) {
		t.Fatalf("expected no telemetry file for nil cfg, err=%v", err)
	}
}

// TestRecord_PreservesPresetTimestamp — a caller-set timestamp is not overwritten.
func TestRecord_PreservesPresetTimestamp(t *testing.T) {
	t.Chdir(t.TempDir())
	fixedNow(t, "2026-06-13T12:00:00Z")
	Record(enabledCfg(), Event{Type: TypeStepAdvanced, Timestamp: "2020-01-01T00:00:00Z"})
	events, err := Read(telemetryDir)
	if err != nil || len(events) != 1 {
		t.Fatalf("read: %v len=%d", err, len(events))
	}
	if events[0].Timestamp != "2020-01-01T00:00:00Z" {
		t.Fatalf("preset timestamp overwritten: %q", events[0].Timestamp)
	}
	if events[0].Schema != Schema {
		t.Fatalf("schema not stamped: %q", events[0].Schema)
	}
}

// TestRecord_IOErrorIsSwallowed — when the log dir cannot be created (a file
// occupies the path), Record warns and returns without panicking.
func TestRecord_IOErrorIsSwallowed(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	// Place a regular file where the .workflow dir must be, so MkdirAll fails.
	if err := os.WriteFile(filepath.Join(dir, ".workflow"), []byte("x"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	Record(enabledCfg(), Event{Type: TypeGateFailure, Gate: "G"}) // must not panic/fail
	// The events file must not be readable: either it does not exist, or the
	// path errors (here .workflow is a file, so stat yields ENOTDIR). Any
	// non-nil stat error means the write was swallowed and nothing was logged.
	if _, err := os.Stat(eventsFile); err == nil {
		t.Fatalf("expected no events file after I/O error, but it exists")
	}
}
