package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/synthesize"
)

// TestRenderInferenceSummary_AmbiguousWithReasons covers the ambiguous note and
// the populated-rationale loop.
func TestRenderInferenceSummary_AmbiguousWithReasons(t *testing.T) {
	inf := synthesize.Inference{
		Best:       synthesize.Hexagonal,
		Confidence: synthesize.Medium,
		Ambiguous:  true,
		Scores: []synthesize.Score{{
			Archetype: synthesize.Hexagonal,
			Signals:   []synthesize.Signal{{Reason: "ports/adapters layout"}},
		}},
	}
	out := RenderInferenceSummary(inf)
	for _, want := range []string{
		"inferred archetype: hexagonal", "confidence: medium",
		"ambiguous", "rationale:", "- ports/adapters layout",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}

// TestRenderInferenceSummary_NoReasons covers the non-ambiguous path and the
// empty-rationale fallback (Best has no matching scored signals).
func TestRenderInferenceSummary_NoReasons(t *testing.T) {
	inf := synthesize.Inference{Best: synthesize.Custom, Confidence: synthesize.Low}
	out := RenderInferenceSummary(inf)
	if !strings.Contains(out, "rationale: (none") {
		t.Fatalf("expected empty rationale fallback: %q", out)
	}
	if strings.Contains(out, "ambiguous") {
		t.Fatalf("non-ambiguous inference should omit the note: %q", out)
	}
}
