package roadmap

import (
	"strings"
	"testing"
)

func phaseStr(p Phase) string { return strings.Join(renderPhase(p), "\n") }

// Phase heading carries an authored status glyph verbatim.
func TestRenderPhaseHeadingGlyph(t *testing.T) {
	got := phaseStr(Phase{Name: "✅ Phase 0: Bootstrap", Features: []Feature{{Name: "a"}}})
	if !strings.HasPrefix(got, "## ✅ Phase 0: Bootstrap\n") {
		t.Fatalf("glyph not preserved in heading: %q", got)
	}
}

// Multi-paragraph note renders as an unbroken blockquote with a bare ">".
func TestRenderPhaseMultiParagraphNote(t *testing.T) {
	got := phaseStr(Phase{Name: "P", Note: "first\n\nsecond", Features: []Feature{{Name: "a"}}})
	want := "## P\n\n> first\n>\n> second\n\n- **a**"
	if got != want {
		t.Fatalf("note mismatch\n got:%q\nwant:%q", got, want)
	}
}

// A phase with no note renders heading then features with no blockquote.
func TestRenderPhaseNoNote(t *testing.T) {
	got := phaseStr(Phase{Name: "P", Features: []Feature{{Name: "a"}}})
	if got != "## P\n\n- **a**" {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(got, ">") {
		t.Fatalf("no blockquote expected: %q", got)
	}
}

// A phase with zero features renders only its heading, no trailing blank.
func TestRenderPhaseZeroFeatures(t *testing.T) {
	got := phaseStr(Phase{Name: "P"})
	if got != "## P" {
		t.Fatalf("empty phase must be just heading, got %q", got)
	}
}

// A noteless empty phase trims the heading's trailing blank.
func TestRenderPhaseEmptyWithNote(t *testing.T) {
	got := phaseStr(Phase{Name: "P", Note: "why"})
	if got != "## P\n\n> why" {
		t.Fatalf("got %q", got)
	}
}

// Backlog phase dispatches to the deferred-finding shape.
func TestRenderPhaseBacklogDispatch(t *testing.T) {
	got := phaseStr(Phase{Name: "Backlog", Features: []Feature{{Name: "x", Summary: "s"}}})
	if got != "## Backlog\n\n- **x** — s" {
		t.Fatalf("got %q", got)
	}
}

// No live status glyph is emitted on feature bullets (only in phase names).
func TestRenderNoFeatureStatusGlyph(t *testing.T) {
	out := mustRender(t, &Roadmap{Phases: []Phase{{Name: "P",
		Features: []Feature{{Name: "a", Description: "done"}}}}})
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "- ") && (strings.Contains(line, "✅") || strings.Contains(line, "✓")) {
			t.Fatalf("feature bullet must not carry a status glyph: %q", line)
		}
	}
}
