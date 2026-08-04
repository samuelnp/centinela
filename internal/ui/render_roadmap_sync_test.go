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
		Reason: "auto-commit is disabled", Regenerated: true, InWorkingTree: true,
	})
	for _, want := range []string{"not committed", "in your working tree", "auto-commit is disabled", "ROADMAP.md regenerated"} {
		if !strings.Contains(got, want) {
			t.Fatalf("%q missing from %q", want, got)
		}
	}
}

func TestRenderRoadmapSyncWarnNamesTheReason(t *testing.T) {
	for _, reason := range []string{"no git repository", "no HEAD", "merge in progress", "rebase in progress", "commit failed: hook rejected"} {
		got := RenderRoadmapSync(roadmap.SyncReport{Warn: true, Reason: reason, InWorkingTree: true})
		if !strings.Contains(got, reason) || !strings.Contains(got, "not committed") {
			t.Fatalf("reason %q not rendered: %q", reason, got)
		}
	}
}

func TestRenderRoadmapSyncNoChangeIsQuiet(t *testing.T) {
	got := RenderRoadmapSync(roadmap.SyncReport{Reason: "roadmap state is unchanged", InWorkingTree: true})
	if strings.Contains(got, "⚠") || !strings.Contains(got, "unchanged") {
		t.Fatalf("an unchanged sync must not warn: %q", got)
	}
	if strings.Contains(got, "regenerated") {
		t.Fatalf("nothing was regenerated: %q", got)
	}
}

// F1: the line an operator reads must never assert the record is on disk when
// the read-back says roadmap.json is no longer what this command wrote.
func TestRenderRoadmapSyncRefusesToClaimAnUnverifiedTree(t *testing.T) {
	got := RenderRoadmapSync(roadmap.SyncReport{
		Reason: "git add failed: index.lock exists", Warn: true,
	})
	if strings.Contains(got, "in your working tree") {
		t.Fatalf("an unverified sync must not claim the record is on disk: %q", got)
	}
	for _, want := range []string{"NOT committed", "no longer matches", "Re-check .workflow/roadmap.json"} {
		if !strings.Contains(got, want) {
			t.Fatalf("%q missing from %q", want, got)
		}
	}
}

// The same refusal applies on the ordinary disable_auto_commit path.
func TestRenderRoadmapSyncUnverifiedOutranksThePolicyNotice(t *testing.T) {
	got := RenderRoadmapSync(roadmap.SyncReport{Reason: "auto-commit is disabled", Regenerated: true})
	if strings.Contains(got, "in your working tree") || !strings.Contains(got, "NOT committed") {
		t.Fatalf("unverified state must outrank the policy notice: %q", got)
	}
}
