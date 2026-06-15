package ui

import (
	"fmt"
	"strings"

	"github.com/samuelnp/centinela/internal/calibration"
)

// RenderCalibration renders the per-model calibration report in house style: a
// header (model count + span), then one block per model citing its class,
// current→recommended profile, verdict, and the raw evidence counts. An empty
// report renders a single "no telemetry yet" line. Models are pre-sorted by
// Calibrate (id asc, "unattributed" last); this never ranges a map. lipgloss
// auto-strips ANSI on non-TTY, so piped output is plain and parseable.
func RenderCalibration(r calibration.Report) string {
	if r.ModelCount == 0 {
		return StyleMuted.Render("no telemetry yet — run governed workflows to populate calibration")
	}
	parts := []string{
		StyleBold.Render(fmt.Sprintf("Calibration — %d models", r.ModelCount)) +
			"  " + StyleMuted.Render(spanLabel(r.SpanStart, r.SpanEnd)),
	}
	for _, m := range r.Models {
		parts = append(parts, modelBlock(m))
	}
	return strings.Join(parts, "\n\n")
}

// modelBlock renders one model's id, profile transition, verdict, and evidence.
func modelBlock(m calibration.ModelRecord) string {
	class := m.Class
	if class == "" {
		class = "(none)"
	}
	header := StyleBold.Render(m.Model) + "  " +
		StyleMuted.Render(fmt.Sprintf("class=%s", class))
	verdict := fmt.Sprintf("  %s  %s",
		string(m.Verdict),
		StyleMuted.Render(calProfileLine(m)))
	return header + "\n" + verdict + "\n" + "  " + StyleMuted.Render(evidenceLine(m.Friction))
}

// profileLine renders "current → recommended (Recommendation)", or just the
// recommendation when there is no profile (Unclassified).
func calProfileLine(m calibration.ModelRecord) string {
	if m.RecommendedProfile == "" {
		return string(m.Recommendation)
	}
	return fmt.Sprintf("%s → %s (%s)", m.CurrentProfile, m.RecommendedProfile, m.Recommendation)
}

// evidenceLine renders the raw counts; rate is "n/a" when HasRate is false.
func evidenceLine(f calibration.FrictionStats) string {
	rate := "n/a"
	if f.HasRate {
		rate = fmt.Sprintf("%.2f", f.Rate)
	}
	return fmt.Sprintf("advances=%d rework=%d rate=%s (blocks=%d gate=%d verify=%d)",
		f.Advances, f.Rework, rate, f.Blocks, f.GateFailures, f.VerifyRejections)
}
