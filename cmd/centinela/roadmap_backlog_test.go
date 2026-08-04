package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

// seedBacklog writes a roadmap whose Backlog holds findings at the given ages
// and returns the captured stdout of runRoadmapBacklog.
func seedBacklog(t *testing.T, ages ...int) {
	t.Helper()
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	feats := make([]string, 0, len(ages))
	for i, d := range ages {
		at := time.Now().AddDate(0, 0, -d).UTC().Format(time.RFC3339)
		feats = append(feats, `{"name":"f`+string(rune('a'+i))+`","summary":"s","deferredAt":"`+at+`"}`)
	}
	body := `{"phases":[{"name":"Phase 1","features":[]},{"name":"Backlog","features":[` +
		strings.Join(feats, ",") + `]}]}`
	if err := os.WriteFile(".workflow/roadmap.json", []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	backlogStaleOnly, backlogOlderThan, backlogJSON = false, 0, false
}

// backlogOut runs the command and captures stdout, failing on any error.
func backlogOut(t *testing.T) string {
	t.Helper()
	var err error
	out := captureStdout(t, func() { err = runRoadmapBacklog(nil, nil) })
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}
	return out
}

func TestRoadmapBacklogListsOldestFirst(t *testing.T) {
	seedBacklog(t, 40, 20, 2)
	out := backlogOut(t)
	if !strings.Contains(out, "3 findings · 2 older than 14d") {
		t.Fatalf("footer wrong:\n%s", out)
	}
	if strings.Index(out, "40d") > strings.Index(out, "2d ") {
		t.Fatalf("not oldest first:\n%s", out)
	}
}

func TestRoadmapBacklogStaleFiltersAndOlderThanOverrides(t *testing.T) {
	seedBacklog(t, 40, 20, 2)
	backlogStaleOnly = true
	out := backlogOut(t)
	if !strings.Contains(out, "40d") || !strings.Contains(out, "20d") || strings.Contains(out, "\n  2d") {
		t.Fatalf("--stale must keep only the two stale rows:\n%s", out)
	}
	backlogOlderThan = 30
	out = backlogOut(t)
	if !strings.Contains(out, "40d") || strings.Contains(out, "20d") {
		t.Fatalf("--older-than 30 must keep only the 40d row:\n%s", out)
	}
	if !strings.Contains(out, "older than 30d") {
		t.Fatalf("the footer must name the overridden threshold:\n%s", out)
	}
}

func TestRoadmapBacklogEmptyIsExitZero(t *testing.T) {
	seedBacklog(t)
	backlogStaleOnly = true
	out := backlogOut(t)
	if !strings.Contains(out, "No deferred findings") {
		t.Fatalf("empty state:\n%s", out)
	}
}
