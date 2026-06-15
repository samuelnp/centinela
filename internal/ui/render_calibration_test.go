package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/calibration"
)

// TestRenderCalibrationEmpty — a zero-model report renders the empty-state line.
func TestRenderCalibrationEmpty(t *testing.T) {
	out := RenderCalibration(calibration.Report{})
	if !strings.Contains(out, "no telemetry yet") {
		t.Fatalf("empty render missing empty-state line: %q", out)
	}
}

// populated builds a two-model report (one classified, one Unclassified).
func populated() calibration.Report {
	return calibration.Report{
		ModelCount: 2,
		SpanStart:  "2026-01-01T00:00:00Z",
		SpanEnd:    "2026-02-01T00:00:00Z",
		Models: []calibration.ModelRecord{
			{
				Model: "claude-sonnet-4-6", Class: "capable", CurrentProfile: "guided",
				RecommendedProfile: "strict", Recommendation: calibration.Tighten,
				Verdict:  calibration.Undergoverned,
				Friction: calibration.FrictionStats{Advances: 3, Rework: 3, Rate: 1.0, HasRate: true, GateFailures: 3},
			},
			{
				Model: "unattributed", Recommendation: calibration.None,
				Verdict:  calibration.Unclassified,
				Friction: calibration.FrictionStats{Advances: 0, Rework: 4, GateFailures: 4},
			},
		},
	}
}

// TestRenderCalibrationPopulated — header, ids, verdicts, profile transition,
// recommendation, and the rate/n-a evidence all appear; deterministic.
func TestRenderCalibrationPopulated(t *testing.T) {
	r := populated()
	out := RenderCalibration(r)
	for _, want := range []string{
		"Calibration — 2 models", "claude-sonnet-4-6", "Undergoverned",
		"guided → strict (Tighten)", "rate=1.00", "unattributed",
		"Unclassified", "class=(none)", "rate=n/a",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("render missing %q:\n%s", want, out)
		}
	}
	if RenderCalibration(r) != out {
		t.Fatal("render not deterministic across runs")
	}
}

// TestRenderCalibrationOrderingPreserved — the renderer emits records in slice
// order (it never re-sorts): sonnet before unattributed.
func TestRenderCalibrationOrderingPreserved(t *testing.T) {
	out := RenderCalibration(populated())
	if strings.Index(out, "claude-sonnet-4-6") > strings.Index(out, "unattributed") {
		t.Fatalf("ordering not preserved:\n%s", out)
	}
}
