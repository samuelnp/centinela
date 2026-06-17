package acceptance_test

import (
	"strings"
	"testing"
)

// Acceptance: specs/roadmap-doc-sync.feature

// Scenario: Backlog phase features render using deferred-finding format
func TestRds_BacklogDeferredFinding(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"Backlog","features":[{"name":"x","summary":"s","deferredAt":"2026-01-01","source":{"feature":"feat","role":"qa"}}]}]}`)
	mustHave(t, out, "- **x** — s *(deferred 2026-01-01 · feat/qa)*")
	if strings.Contains(out, "*Fixes:") || strings.Contains(out, "depends on") {
		t.Fatalf("backlog must not emit Fixes/depends: %q", out)
	}
}

// Scenario: Backlog feature with empty source fields omits the empty parenthetical
func TestRds_BacklogEmptySource(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"Backlog","features":[{"name":"x","summary":"s"}]}]}`)
	mustHave(t, out, "- **x** — s")
	for _, bad := range []string{"()", "· /", "*("} {
		if strings.Contains(out, bad) {
			t.Fatalf("must not contain %q: %q", bad, out)
		}
	}
}

// Scenario: Generated ROADMAP.md contains no per-feature live status glyph
func TestRds_NoFeatureStatusGlyph(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"P","features":[{"name":"a","description":"done"}]},{"name":"Q","features":[{"name":"b"}]}]}`)
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "- ") && (strings.Contains(line, "✅") || strings.Contains(line, "✓")) {
			t.Fatalf("feature bullet carries a status glyph: %q", line)
		}
	}
}

// Scenario: Phase heading status glyphs authored in the phase name are preserved verbatim
func TestRds_PhaseHeadingGlyphPreserved(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"✅ Phase 0: Bootstrap","features":[{"name":"a"}]}]}`)
	mustHave(t, out, "## ✅ Phase 0: Bootstrap")
}

// Scenario: A phase with no features renders only its heading and optional note
func TestRds_PhaseZeroFeatures(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"Empty","note":"why"},{"name":"P","features":[{"name":"a"}]}]}`)
	mustHave(t, out, "## Empty\n\n> why")
	if !strings.HasSuffix(out, "\n") || strings.HasSuffix(out, "\n\n") {
		t.Fatalf("file must remain valid with one trailing newline: %q", out)
	}
}

// Scenario: Non-ASCII characters in prose fields are passed through byte-for-byte
func TestRds_NonASCIIPassthrough(t *testing.T) {
	out := rdsRender(t, `{"phases":[{"name":"P","features":[{"name":"x","description":"café — “quote” éàü"}]}]}`)
	mustHave(t, out, "café — “quote” éàü")
}
