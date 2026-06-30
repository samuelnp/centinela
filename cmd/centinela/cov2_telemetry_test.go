package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeOversizeTelemetry plants a single 2 MiB line (no newline) at the default
// telemetry path. telemetry.Read caps its scanner token at 1 MiB, so the
// scan fails with bufio.ErrTooLong — a deterministic ReadDefault() error that
// drives every command's "telemetry read failed" branch.
func writeOversizeTelemetry(t *testing.T, dir string) {
	t.Helper()
	p := filepath.Join(dir, ".workflow", "telemetry")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	big := strings.Repeat("x", 2*1024*1024)
	if err := os.WriteFile(filepath.Join(p, "events.jsonl"), []byte(big), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCov2DashboardSurfacesTelemetryReadError(t *testing.T) {
	d := chdirIntoTemp(t)
	writeOversizeTelemetry(t, d)
	dashboardJSON = false
	if err := runDashboard(nil, nil); err == nil {
		t.Fatal("expected runDashboard to surface the telemetry read error")
	}
}

func TestCov2InsightsSurfacesTelemetryReadError(t *testing.T) {
	d := chdirIntoTemp(t)
	writeOversizeTelemetry(t, d)
	insightsJSON, insightsTop = false, 5
	if err := runInsights(nil, nil); err == nil {
		t.Fatal("expected runInsights to surface the telemetry read error")
	}
}

func TestCov2CostSurfacesTelemetryReadError(t *testing.T) {
	d := chdirIntoTemp(t)
	writeOversizeTelemetry(t, d)
	costJSON = false
	if err := runCost(nil, nil); err == nil {
		t.Fatal("expected runCost to surface the telemetry read error")
	}
}

func TestCov2CalibrateSurfacesTelemetryReadError(t *testing.T) {
	d := chdirIntoTemp(t)
	writeOversizeTelemetry(t, d)
	calibrateJSON = false
	if err := runCalibrate(nil, nil); err == nil {
		t.Fatal("expected runCalibrate to surface the telemetry read error")
	}
}
