package roadmap

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func deferSetup(t *testing.T, src string) (string, string) {
	t.Helper()
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(src), 0644) //nolint:errcheck
	return d, p
}

// TestDefer_HappyPath creates a Backlog phase and appends the finding.
func TestDefer_HappyPath(t *testing.T) {
	_, p := deferSetup(t, minimalRoadmapJSON)
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	opts := DeferOptions{
		Slug:    "hook-timeout-config",
		Summary: "Prewrite hook timeout is hardcoded",
		Source:  &Source{Feature: "deferred-findings-roadmap-capture", Role: "senior-engineer"},
		Now:     ts,
	}
	if err := Defer(p, opts); err != nil {
		t.Fatalf("Defer: %v", err)
	}
	data, _ := os.ReadFile(p)
	s := string(data)
	if !strings.Contains(s, "Backlog") {
		t.Error("Backlog phase missing")
	}
	if !strings.Contains(s, "hook-timeout-config") {
		t.Error("slug missing in output")
	}
	if !strings.Contains(s, "2026-01-01T00:00:00Z") {
		t.Error("deferredAt timestamp missing")
	}
}

// TestDefer_PreservesExistingEntries verifies byte-stable prior entries.
func TestDefer_PreservesExistingEntries(t *testing.T) {
	src := `{"phases":[{"name":"Phase 0","features":[{"name":"f1","customField":"keep-me"}]}]}`
	_, p := deferSetup(t, src)
	Defer(p, DeferOptions{Slug: "new-finding", Summary: "s", Now: time.Now()}) //nolint:errcheck
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), "keep-me") {
		t.Error("custom field in existing entry must be preserved")
	}
	if !strings.Contains(string(data), "f1") {
		t.Error("existing f1 feature must still be present")
	}
}

// TestDefer_EmptySummary is rejected before any write (regression: summary validation).
func TestDefer_EmptySummary(t *testing.T) {
	_, p := deferSetup(t, minimalRoadmapJSON)
	before, _ := os.ReadFile(p)
	err := Defer(p, DeferOptions{Slug: "empty-summary-test", Summary: ""})
	if err == nil {
		t.Fatal("empty summary must be rejected")
	}
	if !strings.Contains(err.Error(), "summary") {
		t.Errorf("error should mention summary, got: %v", err)
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Error("roadmap.json must be unchanged on empty summary")
	}
}

// TestDefer_WhitespaceSummary rejected before any write.
func TestDefer_WhitespaceSummary(t *testing.T) {
	_, p := deferSetup(t, minimalRoadmapJSON)
	before, _ := os.ReadFile(p)
	if err := Defer(p, DeferOptions{Slug: "ws-test", Summary: "   "}); err == nil {
		t.Fatal("whitespace summary must be rejected")
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Error("roadmap.json must be unchanged")
	}
}
