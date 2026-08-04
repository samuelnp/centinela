package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func TestRenderRoadmapSyncCommitted(t *testing.T) {
	got := RenderRoadmapSync(roadmap.SyncReport{
		Message:   "chore(roadmap): defer flaky-thing",
		Paths:     []string{".workflow/roadmap.json", "ROADMAP.md"},
		Committed: true, Regenerated: true,
	})
	for _, want := range []string{"Committed roadmap state", "chore(roadmap): defer flaky-thing", ".workflow/roadmap.json", "ROADMAP.md"} {
		if !strings.Contains(got, want) {
			t.Fatalf("%q missing from %q", want, got)
		}
	}
}

func TestRenderRoadmapSyncDisabledSaysUncommitted(t *testing.T) {
	got := RenderRoadmapSync(roadmap.SyncReport{
		Reason: "auto-commit is disabled", Regenerated: true,
	})
	for _, want := range []string{"left uncommitted", "auto-commit is disabled", "ROADMAP.md regenerated"} {
		if !strings.Contains(got, want) {
			t.Fatalf("%q missing from %q", want, got)
		}
	}
}

func TestRenderRoadmapSyncWarnNamesTheReason(t *testing.T) {
	for _, reason := range []string{"no git repository", "no HEAD", "merge in progress", "rebase in progress", "commit failed: hook rejected"} {
		got := RenderRoadmapSync(roadmap.SyncReport{Warn: true, Reason: reason})
		if !strings.Contains(got, reason) || !strings.Contains(got, "left uncommitted") {
			t.Fatalf("reason %q not rendered: %q", reason, got)
		}
	}
}

func TestRenderRoadmapSyncNoChangeIsQuiet(t *testing.T) {
	got := RenderRoadmapSync(roadmap.SyncReport{Reason: "roadmap state is unchanged"})
	if strings.Contains(got, "⚠") || !strings.Contains(got, "unchanged") {
		t.Fatalf("an unchanged sync must not warn: %q", got)
	}
	if strings.Contains(got, "regenerated") {
		t.Fatalf("nothing was regenerated: %q", got)
	}
}
