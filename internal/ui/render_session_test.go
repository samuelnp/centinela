package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/workflow"
)

func rmapUI(features ...string) *roadmap.Roadmap {
	p := roadmap.Phase{Name: "Phase 0"}
	for _, f := range features {
		p.Features = append(p.Features, roadmap.Feature{Name: f})
	}
	return &roadmap.Roadmap{Phases: []roadmap.Phase{p}}
}

// ready branch: banner + roadmap body + ready list + pointer PATHS, no inlined contents.
func TestRenderSessionRehydration_HasNext(t *testing.T) {
	out := RenderSessionRehydration(rmapUI("next-feature"), []string{"next-feature"}, true)
	for _, want := range []string{
		"rehydration", "next-feature",
		"Ready to start now:",
		"PROJECT.md", "docs/features/next-feature.md",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in payload:\n%s", want, out)
		}
	}
	if strings.Contains(out, "## Problem") {
		t.Fatalf("payload must not inline brief contents:\n%s", out)
	}
}

// !hasIncomplete branch: roadmap-complete line, no pointer.
func TestRenderSessionRehydration_NoNext(t *testing.T) {
	out := RenderSessionRehydration(rmapUI("done-a"), nil, false)
	if !strings.Contains(out, "Roadmap complete") {
		t.Fatalf("expected roadmap-complete line:\n%s", out)
	}
	if strings.Contains(out, "docs/features/") {
		t.Fatalf("complete payload must not emit a <next>.md pointer:\n%s", out)
	}
	if !strings.Contains(out, "PROJECT.md") {
		t.Fatalf("PROJECT.md pointer should still be present:\n%s", out)
	}
}

// RenderContextCapped: more>0 appends "+N more"; more==0 omits it.
func TestRenderContextCapped_MoreBranch(t *testing.T) {
	wfs := []*workflow.Workflow{workflow.New("a"), workflow.New("b")}
	if !strings.Contains(RenderContextCapped(wfs, 3), "+3 more") {
		t.Fatal("expected '+3 more' when more>0")
	}
	if strings.Contains(RenderContextCapped(wfs, 0), "more active") {
		t.Fatal("expected no '+N more' hint when more==0")
	}
}
