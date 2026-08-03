package orchestration

import "testing"

func TestFillSlotRendersTheCanonicalMarker(t *testing.T) {
	if got := FillSlot("the impl file path"); got != "<FILL: the impl file path>" {
		t.Fatalf("FillSlot = %q", got)
	}
}

// UnreplacedSlot is the single rule separating a scaffold from something an
// author wrote. It must fire on a real slot ANYWHERE in the text — position is
// not part of the question — and must NOT fire on the generic citation form,
// which is how the CLI's own errors and docs refer to the marker.
func TestUnreplacedSlotSeparatesSlotsFromCitations(t *testing.T) {
	for text, want := range map[string]bool{
		FillSlot("type"):                     true,
		"- " + FillSlot("type") + ": tidy":   true,
		"## Changelog":                       false,
		"- fix: <FILL: one-line summary>":    true,
		"- fix: quote the <FILL: ...> form":  false,
		"- fix: quote the <FILL: …> form":    false,
		"- fix: an empty <FILL: > slot":      false,
		"- docs: <FILL: without a closer":    false,
		"- feat: bind the gates":             false,
		"":                                   false,
		"<FILL: a>x<FILL: ...>":              true,
		"<FILL: ...> then <FILL: real slot>": true,
		"<fill: lowercase>":                  false,
		"< FILL: spaced>":                    false,
	} {
		if got := UnreplacedSlot(text); got != want {
			t.Errorf("UnreplacedSlot(%q) = %v, want %v", text, got, want)
		}
	}
}

// Every artifact body the CLI scaffolds must be detectable as a scaffold by
// the same predicate the gates use — the renderer and the detector share one
// marker, and this is what keeps them from drifting apart.
func TestScaffoldedSlotsAreDetectable(t *testing.T) {
	for _, desc := range []string{"type", "one-line summary of the change", "each file checked"} {
		if !UnreplacedSlot("- " + FillSlot(desc)) {
			t.Errorf("FillSlot(%q) must be detected as unreplaced", desc)
		}
	}
}
