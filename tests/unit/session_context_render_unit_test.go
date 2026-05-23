package unit_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
	"github.com/samuelnp/centinela/internal/ui"
	"github.com/samuelnp/centinela/internal/workflow"
)

// rmap builds a roadmap with a single phase holding the named features.
func rmap(features ...string) *roadmap.Roadmap {
	p := roadmap.Phase{Name: "Phase 0"}
	for _, f := range features {
		p.Features = append(p.Features, roadmap.Feature{Name: f})
	}
	return &roadmap.Roadmap{Phases: []roadmap.Phase{p}}
}

// Spec Half-B: the rehydration payload carries the banner, the roadmap body with
// per-feature status, the next-feature reference and the two pointer PATHS — and
// never inlines file contents (paths only).
func TestRenderSessionRehydration_SuccessPayloadHasPointersNotContents(t *testing.T) {
	r := rmap("next-feature", "later-feature")
	out := ui.RenderSessionRehydration(r, "next-feature", true)
	for _, want := range []string{
		"rehydration",                        // banner
		"next-feature",                       // roadmap body feature
		"(planned)",                          // per-feature status from RenderRoadmap
		"Next feature to plan: next-feature", // next reference
		"PROJECT.md",                         // pointer path
		"docs/features/next-feature.md",      // pointer path
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected payload to contain %q, got:\n%s", want, out)
		}
	}
	// Paths only: the banner/body must not inline file bodies. A literal markdown
	// heading like "## Problem" would only appear if a brief were inlined.
	if strings.Contains(out, "## Problem") || strings.Contains(out, "## User Stories") {
		t.Fatalf("payload inlined file contents, expected paths only:\n%s", out)
	}
}

// Spec roadmap-complete: !hasNext renders a complete-style line and emits NO
// docs/features/<next>.md pointer.
func TestRenderSessionRehydration_CompleteHasNoNextPointer(t *testing.T) {
	r := rmap("done-a")
	out := ui.RenderSessionRehydration(r, "", false)
	if !strings.Contains(out, "Roadmap complete") {
		t.Fatalf("expected roadmap-complete line, got:\n%s", out)
	}
	if strings.Contains(out, "docs/features/") {
		t.Fatalf("complete payload must not emit a <next>.md pointer, got:\n%s", out)
	}
	if !strings.Contains(out, "PROJECT.md") {
		t.Fatalf("PROJECT.md pointer should still be present, got:\n%s", out)
	}
}

// Spec scenarios 5/6: the capped panel appends a "+N more" hint when more>0 and
// omits any "+N more" text when more==0.
func TestRenderContextCapped_MoreHintBranches(t *testing.T) {
	wfs := []*workflow.Workflow{workflow.New("a"), workflow.New("b")}
	withMore := ui.RenderContextCapped(wfs, 2)
	if !strings.Contains(withMore, "+2 more") {
		t.Fatalf("expected '+2 more' hint when more>0, got:\n%s", withMore)
	}
	noMore := ui.RenderContextCapped(wfs, 0)
	if strings.Contains(noMore, "more active") || strings.Contains(noMore, "+") {
		t.Fatalf("expected no '+N more' hint when more==0, got:\n%s", noMore)
	}
}
