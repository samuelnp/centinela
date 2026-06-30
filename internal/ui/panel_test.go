package ui

import (
	"strings"
	"testing"
)

const borderChars = "╭╮╰╯│─"

func hasBorder(s string) bool {
	return strings.ContainsAny(s, borderChars)
}

func TestRenderSystemPanelHasNoBorder(t *testing.T) {
	out := renderSystemPanel("ROADMAP", "PHASE OVERVIEW", toneInfo, "Phase 0: Bootstrap")
	if hasBorder(out) {
		t.Fatalf("panel should have no border, got:\n%s", out)
	}
	for _, want := range []string{"ROADMAP", "PHASE OVERVIEW", "Phase 0: Bootstrap", "🛡️👁️"} {
		if !strings.Contains(out, want) {
			t.Errorf("panel lost content %q", want)
		}
	}
}

func TestRenderSystemPanelKeepsHeaderThenBody(t *testing.T) {
	out := renderSystemPanel("HOOK", "TITLE", toneWarn, "body line")
	// header line first, blank separator, then body — no box framing.
	if !strings.HasPrefix(strings.TrimLeft(out, " "), "🛡️👁️") {
		t.Fatalf("expected the persona header first, got:\n%s", out)
	}
	if !strings.Contains(out, "\n\nbody line") {
		t.Errorf("expected a blank line before the body, got:\n%q", out)
	}
}

func TestRenderBlockedHasNoBorder(t *testing.T) {
	out := RenderBlocked("code", "plan", "f", "/tmp/a.go")
	if hasBorder(out) {
		t.Fatalf("blocked-write directive should have no border, got:\n%s", out)
	}
	if !strings.Contains(out, "BLOCKED WRITE") || !strings.Contains(out, "Next action") {
		t.Error("blocked-write directive lost its content")
	}
}
