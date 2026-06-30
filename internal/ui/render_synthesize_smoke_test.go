package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/synthesize"
)

func TestRenderInferenceSummary_ZeroValue(t *testing.T) {
	out := RenderInferenceSummary(synthesize.Inference{})
	if !strings.Contains(out, "inferred archetype") {
		t.Fatalf("missing archetype line, got %q", out)
	}
	if !strings.Contains(out, "(none") {
		t.Fatalf("want none-rationale for zero scores, got %q", out)
	}
}

func TestRenderInferenceSummary_WithRationale(t *testing.T) {
	inf := synthesize.Inference{
		Best:       "hexagonal",
		Confidence: "high",
		Scores: []synthesize.Score{{
			Archetype: "hexagonal",
			Total:     10,
			Signals:   []synthesize.Signal{{Reason: "has internal/domain", Weight: 5}},
		}},
	}
	out := RenderInferenceSummary(inf)
	if !strings.Contains(out, "hexagonal") {
		t.Fatalf("archetype missing, got %q", out)
	}
	if !strings.Contains(out, "high") {
		t.Fatalf("confidence missing, got %q", out)
	}
	if !strings.Contains(out, "has internal/domain") {
		t.Fatalf("rationale missing, got %q", out)
	}
}

func TestRenderInferenceSummary_AmbiguousNote(t *testing.T) {
	inf := synthesize.Inference{
		Best:      "hexagonal",
		Ambiguous: true,
		Scores: []synthesize.Score{{
			Archetype: "hexagonal",
			Signals:   []synthesize.Signal{{Reason: "dirs", Weight: 3}},
		}},
	}
	out := RenderInferenceSummary(inf)
	if !strings.Contains(out, "ambiguous") {
		t.Fatalf("ambiguous note missing, got %q", out)
	}
}

func TestRenderInferenceSummary_NoTrailingNewline(t *testing.T) {
	out := RenderInferenceSummary(synthesize.Inference{Best: "n-tier", Confidence: "low"})
	if strings.HasSuffix(out, "\n") {
		t.Fatalf("output must not end with newline, got %q", out)
	}
}
