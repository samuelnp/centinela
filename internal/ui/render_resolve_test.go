package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func TestRenderResolveSummaryNamesEachSidesContribution(t *testing.T) {
	got := RenderResolveSummary(roadmap.Merged{Kept: 10, FromBase: 5, FromOurs: 2, FromTheirs: 3})
	for _, want := range []string{"kept 10 findings", "5 unchanged", "2 from our side", "3 from theirs"} {
		if !strings.Contains(got, want) {
			t.Fatalf("%q missing from %q", want, got)
		}
	}
}

func TestRenderResolveOtherStates(t *testing.T) {
	if !strings.Contains(RenderResolveMarkdownOnly(), "Regenerated ROADMAP.md") {
		t.Fatal("the markdown-only state must say what it did")
	}
	if !strings.Contains(RenderResolveNothingToDo(), "Nothing to resolve") {
		t.Fatal("the no-op state must be explicit")
	}
}
