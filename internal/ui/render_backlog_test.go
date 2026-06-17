package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

// TestRenderBacklogSection_WithFindings renders slug and summary lines.
func TestRenderBacklogSection_WithFindings(t *testing.T) {
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{
		{Name: "Backlog", Features: []roadmap.Feature{
			{Name: "hook-timeout-config", Summary: "Timeout hardcoded"},
			{Name: "doc-sync-reminder", Summary: "Docs stale"},
		}},
	}}
	got := renderBacklogSection(r)
	if !strings.Contains(got, "hook-timeout-config") {
		t.Error("slug 1 missing")
	}
	if !strings.Contains(got, "doc-sync-reminder") {
		t.Error("slug 2 missing")
	}
	if !strings.Contains(got, "Timeout hardcoded") {
		t.Error("summary 1 missing")
	}
	if !strings.Contains(got, roadmap.BacklogPhaseName) {
		t.Error("Backlog heading missing")
	}
}

// TestRenderBacklogSection_NoBacklog returns empty string.
func TestRenderBacklogSection_NoBacklog(t *testing.T) {
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{
		{Name: "Phase 0", Features: []roadmap.Feature{{Name: "real"}}},
	}}
	if got := renderBacklogSection(r); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// TestRenderBacklogSection_EmptyBacklog returns empty string.
func TestRenderBacklogSection_EmptyBacklog(t *testing.T) {
	r := &roadmap.Roadmap{Phases: []roadmap.Phase{
		{Name: "Backlog", Features: []roadmap.Feature{}},
	}}
	if got := renderBacklogSection(r); got != "" {
		t.Errorf("expected empty string for empty Backlog, got %q", got)
	}
}
