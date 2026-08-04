package ui

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/roadmap"
)

func agedRow(name string, days int, known bool, stale bool) roadmap.Aged {
	return roadmap.Aged{
		Feature: roadmap.Feature{
			Name: name, Summary: "a one line summary",
			Source: &roadmap.Source{Feature: "feat", Role: "eng"},
		},
		AgeDays: days, KnownAge: known, Stale: stale,
	}
}

func TestRenderBacklogListShowsAgeSlugSourceAndSummary(t *testing.T) {
	rows := []roadmap.Aged{agedRow("old-thing", 40, true, true), agedRow("young", 2, true, false)}
	stats := roadmap.BacklogStats{Total: 2, Stale: 1, ThresholdDays: 14, OldestDays: 40, OldestSlug: "old-thing", OldestKnown: true}
	got := RenderBacklogList(rows, stats, false)
	for _, want := range []string{"40d", "old-thing", "feat/eng", "a one line summary", "2d", "young"} {
		if !strings.Contains(got, want) {
			t.Fatalf("%q missing from:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "2 findings · 1 older than 14d · oldest 40d (old-thing)") {
		t.Fatalf("footer arithmetic missing:\n%s", got)
	}
	if strings.Index(got, "old-thing") > strings.Index(got, "young") {
		t.Fatal("rows must stay oldest-first")
	}
}

func TestRenderBacklogListUnknownAge(t *testing.T) {
	rows := []roadmap.Aged{agedRow("legacy", 0, false, true)}
	got := RenderBacklogList(rows, roadmap.BacklogStats{Total: 1, Stale: 1, ThresholdDays: 14}, false)
	if !strings.Contains(got, "unknown") {
		t.Fatalf("an unreadable clock must render as unknown:\n%s", got)
	}
	if strings.Contains(got, "oldest") {
		t.Fatalf("no oldest may be claimed:\n%s", got)
	}
}

// E7 vs E8: two different empty states, and neither is an error.
func TestRenderBacklogListEmptyStates(t *testing.T) {
	empty := RenderBacklogList(nil, roadmap.BacklogStats{ThresholdDays: 14}, false)
	if !strings.Contains(empty, "No deferred findings") {
		t.Fatalf("empty Backlog: %q", empty)
	}
	nothingStale := RenderBacklogList(nil, roadmap.BacklogStats{Total: 6, ThresholdDays: 14}, true)
	if !strings.Contains(nothingStale, "No findings older than 14d") || !strings.Contains(nothingStale, "6 in the Backlog") {
		t.Fatalf("filtered empty: %q", nothingStale)
	}
}

func TestRenderBacklogListTruncatesLongSummaries(t *testing.T) {
	row := agedRow("x", 1, true, false)
	row.Summary = strings.Repeat("word ", 60)
	got := RenderBacklogList([]roadmap.Aged{row}, roadmap.BacklogStats{Total: 1, ThresholdDays: 14}, false)
	if !strings.Contains(got, "…") {
		t.Fatalf("a long summary must be truncated:\n%s", got)
	}
	for _, line := range strings.Split(got, "\n") {
		if len([]rune(line)) > 140 {
			t.Fatalf("row is %d runes, too wide: %q", len([]rune(line)), line)
		}
	}
}

func TestRenderBacklogListWithoutASource(t *testing.T) {
	row := agedRow("x", 1, true, false)
	row.Source = nil
	if !strings.Contains(RenderBacklogList([]roadmap.Aged{row}, roadmap.BacklogStats{Total: 1, ThresholdDays: 14}, false), "—") {
		t.Fatal("a finding with no provenance must render a placeholder")
	}
	row.Source = &roadmap.Source{Feature: "feat"}
	if strings.Contains(RenderBacklogList([]roadmap.Aged{row}, roadmap.BacklogStats{Total: 1, ThresholdDays: 14}, false), "feat/") {
		t.Fatal("an empty role must not render a trailing slash")
	}
}

func TestRenderBacklogNudgeNamesBothCommands(t *testing.T) {
	got := RenderBacklogNudge(roadmap.Nudge{Total: 12, Stale: 9, ThresholdDays: 14})
	for _, want := range []string{"Roadmap complete", "12 deferred findings", "9 older than 14d",
		"centinela roadmap backlog --stale", "centinela roadmap promote"} {
		if !strings.Contains(got, want) {
			t.Fatalf("%q missing from:\n%s", want, got)
		}
	}
}
