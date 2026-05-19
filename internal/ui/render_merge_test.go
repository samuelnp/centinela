package ui

import (
	"strings"
	"testing"
)

func TestRenderMergeStewardNeeded_NamesFeatureReasonAndResume(t *testing.T) {
	out := RenderMergeStewardNeeded("delta", "git-text-conflict")
	for _, want := range []string{
		"delta", "git-text-conflict",
		"merge-steward-prompt.md", "centinela merge --continue delta",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderMergeStewardNeeded missing %q:\n%s", want, out)
		}
	}
}

func TestRenderMergeEscalated_NamesFeatureAndKeepsContext(t *testing.T) {
	out := RenderMergeEscalated("kappa")
	for _, want := range []string{"kappa", "escalated", "manual review"} {
		if !strings.Contains(strings.ToLower(out), want) {
			t.Fatalf("RenderMergeEscalated missing %q:\n%s", want, out)
		}
	}
}
