package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/brownmap"
)

func TestRenderBrownfieldSummary_CountsAndPath(t *testing.T) {
	out := RenderBrownfieldSummary(brownmap.Plan{BaselineCount: 3, GapCount: 2, DraftPath: ".workflow/d.json"})
	for _, want := range []string{"baseline entries: 3", "gaps: 2", "draft written: .workflow/d.json"} {
		if !strings.Contains(out, want) {
			t.Fatalf("summary missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "supply --goal") {
		t.Fatal("no-gaps hint must NOT appear when there are gaps")
	}
}

func TestRenderBrownfieldSummary_NoGapsHint(t *testing.T) {
	out := RenderBrownfieldSummary(brownmap.Plan{BaselineCount: 1, GapCount: 0, DraftPath: ".workflow/d.json"})
	if !strings.Contains(out, "supply --goal") {
		t.Fatalf("zero gaps must hint at --goal:\n%s", out)
	}
}
