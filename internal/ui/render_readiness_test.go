package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// roadmapIcon maps every status to its icon, including the ready/blocked and
// default branches not reached elsewhere.
func TestRoadmapIcon_AllBranches(t *testing.T) {
	cases := map[string]string{
		"done":        IconDone,
		"in-progress": IconActive,
		"ready":       IconReady,
		"blocked":     IconBlocked,
		"planned":     IconPending,
		"weird":       IconPending,
	}
	for status, want := range cases {
		if got := roadmapIcon(status); got != want {
			t.Fatalf("roadmapIcon(%q) = %q, want %q", status, got, want)
		}
	}
}

// readinessMarker returns the icon + annotation for each state; blocked joins the
// BlockedBy names into the annotation; an unknown state falls back to planned.
func TestReadinessMarker_IconAndAnnotation(t *testing.T) {
	icon, ann := readinessMarker(roadmap.FeatureReadiness{State: "ready"})
	if icon != IconReady || ann != "(ready)" {
		t.Fatalf("ready marker: %q %q", icon, ann)
	}
	icon, ann = readinessMarker(roadmap.FeatureReadiness{State: "blocked", BlockedBy: []string{"a", "b"}})
	if icon != IconBlocked || !strings.Contains(ann, "a, b") {
		t.Fatalf("blocked marker should join names: %q %q", icon, ann)
	}
	if icon, ann := readinessMarker(roadmap.FeatureReadiness{State: "done"}); icon != IconDone || ann != "(done)" {
		t.Fatalf("done marker: %q %q", icon, ann)
	}
	if icon, ann := readinessMarker(roadmap.FeatureReadiness{State: "in-progress"}); icon != IconActive || ann != "(in-progress)" {
		t.Fatalf("in-progress marker: %q %q", icon, ann)
	}
	if icon, ann := readinessMarker(roadmap.FeatureReadiness{State: "???"}); icon != IconPending || ann != "(planned)" {
		t.Fatalf("default marker: %q %q", icon, ann)
	}
}

// RenderReadyList prints 🔓 + each name when populated; a non-empty muted
// empty-state (no names, no ready icon) when empty.
func TestRenderReadyList_Branches(t *testing.T) {
	out := RenderReadyList([]string{"feature-b", "feature-c"})
	if !strings.Contains(out, IconReady) || !strings.Contains(out, "feature-b") || !strings.Contains(out, "feature-c") {
		t.Fatalf("populated list should show ready icon + names:\n%s", out)
	}
	empty := RenderReadyList(nil)
	if strings.TrimSpace(empty) == "" || strings.Contains(empty, IconReady) {
		t.Fatalf("empty list must render a non-empty line with no ready icon:\n%s", empty)
	}
}

// renderReadyBlock: ready frontier header when ready>0; complete line when empty
// and nothing incomplete; "in-progress or blocked" line when empty but work remains.
func TestRenderReadyBlock_Branches(t *testing.T) {
	if !strings.Contains(renderReadyBlock([]string{"f"}, true), "Ready to start now:") {
		t.Fatal("populated frontier should show the ready header")
	}
	if !strings.Contains(renderReadyBlock(nil, false), "Roadmap complete") {
		t.Fatal("empty + complete should show roadmap-complete")
	}
	blocked := renderReadyBlock(nil, true)
	if !strings.Contains(blocked, "in-progress or blocked") {
		t.Fatalf("empty + incomplete should explain the block:\n%s", blocked)
	}
}
