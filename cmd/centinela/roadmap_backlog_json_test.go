package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestRoadmapBacklogJSONShape(t *testing.T) {
	seedBacklog(t, 40, 2)
	backlogJSON = true
	out := backlogOut(t)
	var got backlogPayload
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if got.ThresholdDays != 14 || got.Total != 2 || got.Stale != 1 || len(got.Findings) != 2 {
		t.Fatalf("payload = %+v", got)
	}
	first := got.Findings[0]
	if first.Slug == "" || first.DeferredAt == "" || first.AgeDays == nil || *first.AgeDays != 40 || !first.Stale {
		t.Fatalf("finding = %+v", first)
	}
	for _, key := range []string{`"threshold_days"`, `"total"`, `"stale"`, `"findings"`, `"ageDays"`, `"source"`} {
		if !strings.Contains(out, key) {
			t.Fatalf("%s missing from payload:\n%s", key, out)
		}
	}
}

func TestRoadmapBacklogJSONNullsAnUnknownAge(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"phases":[{"name":"Backlog","features":[{"name":"legacy","summary":"s"}]}]}`
	if err := os.WriteFile(".workflow/roadmap.json", []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	backlogStaleOnly, backlogOlderThan, backlogJSON = true, 0, true
	out := backlogOut(t)
	if !strings.Contains(out, `"ageDays": null`) || !strings.Contains(out, `"stale": true`) {
		t.Fatalf("an unknown age must be null and stale:\n%s", out)
	}
}

// A hint must never break `complete`: an unreadable roadmap is simply silent.
func TestPrintBacklogNudgeSurvivesAnUnreadableRoadmap(t *testing.T) {
	t.Chdir(t.TempDir())
	printBacklogNudge()
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".workflow/roadmap.json", []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out := captureStdout(t, printBacklogNudge); out != "" {
		t.Fatalf("an unparseable roadmap must print nothing: %q", out)
	}
}

func TestRoadmapBacklogSurfacesAMissingRoadmap(t *testing.T) {
	t.Chdir(t.TempDir())
	backlogStaleOnly, backlogOlderThan, backlogJSON = false, 0, false
	if err := runRoadmapBacklog(nil, nil); err == nil {
		t.Fatal("a missing roadmap.json must be a command error")
	}
}

// The positive nudge path: the roadmap is finished and the Backlog is not.
func TestPrintBacklogNudgeFiresOnAFinishedRoadmap(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll(".workflow", 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"phases":[{"name":"Backlog","features":[` +
		`{"name":"old","summary":"s","deferredAt":"2020-01-01T00:00:00Z"}]}]}`
	if err := os.WriteFile(".workflow/roadmap.json", []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, printBacklogNudge)
	for _, want := range []string{"Roadmap complete", "1 deferred findings",
		"centinela roadmap backlog --stale", "centinela roadmap promote"} {
		if !strings.Contains(out, want) {
			t.Fatalf("%q missing from %q", want, out)
		}
	}
}
