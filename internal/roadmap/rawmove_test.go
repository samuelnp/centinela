package roadmap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func roadmapWithBacklog(t *testing.T, slug string) (string, string) {
	t.Helper()
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	src := `{"phases":[{"name":"Phase 5","features":[{"name":"existing"}]},` +
		`{"name":"Backlog","features":[{"name":"` + slug + `","summary":"s","deferredAt":"2026-01-01T00:00:00Z"}]}]}`
	os.WriteFile(p, []byte(src), 0644) //nolint:errcheck
	return d, p
}

// TestFindInBacklog_Found returns raw bytes and index.
func TestFindInBacklog_Found(t *testing.T) {
	_, p := roadmapWithBacklog(t, "my-finding")
	doc, _ := readRawRoadmap(p)
	raw, idx, err := doc.findInBacklog("my-finding")
	if err != nil || idx < 0 || len(raw) == 0 {
		t.Fatalf("findInBacklog: raw=%s idx=%d err=%v", raw, idx, err)
	}
}

// TestFindInBacklog_NotFound returns an error.
func TestFindInBacklog_NotFound(t *testing.T) {
	_, p := roadmapWithBacklog(t, "my-finding")
	doc, _ := readRawRoadmap(p)
	if _, _, err := doc.findInBacklog("nonexistent"); err == nil {
		t.Error("expected error for missing Backlog finding")
	}
}

// TestFindInBacklog_NoBacklogPhase returns an error.
func TestFindInBacklog_NoBacklogPhase(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "roadmap.json")
	os.WriteFile(p, []byte(minimalRoadmapJSON), 0644) //nolint:errcheck
	doc, _ := readRawRoadmap(p)
	if _, _, err := doc.findInBacklog("x"); err == nil {
		t.Error("expected error when no Backlog phase")
	}
}

// TestRemoveBacklogFeature drops the named slug.
func TestRemoveBacklogFeature(t *testing.T) {
	_, p := roadmapWithBacklog(t, "my-finding")
	doc, _ := readRawRoadmap(p)
	_, idx, _ := doc.findInBacklog("my-finding")
	if err := doc.removeBacklogFeature(idx, "my-finding"); err != nil {
		t.Fatalf("removeBacklogFeature: %v", err)
	}
	writeRawRoadmap(p, doc) //nolint:errcheck
	data, _ := os.ReadFile(p)
	if strings.Contains(string(data), "my-finding") {
		t.Error("slug must be removed from Backlog")
	}
}
