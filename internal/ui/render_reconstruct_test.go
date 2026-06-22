package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/reconstruct"
)

func TestRenderReconstructionSummary(t *testing.T) {
	r := reconstruct.Reconstruction{
		Targets:   []reconstruct.Target{{Slug: "a"}, {Slug: "b"}},
		Written:   []string{"specs/a.feature", "features/a.md"},
		Skipped:   []string{"b"},
		TodoCount: 6,
	}
	out := RenderReconstructionSummary(r)
	for _, want := range []string{
		"targets selected: 2",
		"files written: 2",
		"files skipped: 1",
		"skipped (hand-authored spec exists): b",
		"TODO confirm markers: 6",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("summary missing %q:\n%s", want, out)
		}
	}
}

func TestRenderReconstructionSummary_EmptyNoSkips(t *testing.T) {
	out := RenderReconstructionSummary(reconstruct.Reconstruction{})
	if !strings.Contains(out, "targets selected: 0") || strings.Contains(out, "skipped (hand-authored") {
		t.Fatalf("empty summary wrong:\n%s", out)
	}
}
