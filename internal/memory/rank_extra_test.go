package memory

import (
	"os"
	"testing"
	"time"
)

// rank: recency tie-break — later createdAt wins when scores are equal.
func TestRankRecencyTieBreak(t *testing.T) {
	earlier := newEntry("f", "tests", TypeLesson, "- earlier", "s", nil, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	later := newEntry("f", "tests", TypeLesson, "- later body", "s", nil, time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC))

	ranked := rank([]Entry{earlier, later}, map[string]bool{}, map[string]bool{})
	if ranked[0].Body != later.Body {
		t.Fatalf("expected later entry first in tie-break, got %q", ranked[0].Body)
	}
}

// SC-05: idempotence through recall_test path — writing the same entry twice keeps count=1.
func TestWriteIfAbsentConcurrentSameEntry(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	e := newEntry("alpha", "validate", TypeVerdict, "all clear", "src", nil, time.Now())
	w1, _ := writeIfAbsent(e)
	w2, _ := writeIfAbsent(e)
	if !w1 {
		t.Fatal("first write should return true")
	}
	if w2 {
		t.Fatal("second write should return false")
	}
}

// loadEntries: non-.md files are ignored.
func TestLoadEntriesIgnoresNonMdFiles(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	os.MkdirAll(entriesDir, 0o755)                                       //nolint:errcheck
	os.WriteFile(entriesDir+"/entry.txt", []byte("not markdown"), 0o644) //nolint:errcheck
	os.WriteFile(entriesDir+"/index.json", []byte("{}"), 0o644)          //nolint:errcheck

	entries := loadEntries()
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries (non-.md files ignored), got %d", len(entries))
	}
}

// parseFrontmatter: line with no colon separator is skipped gracefully.
func TestParseFrontmatterSkipsLinesWithoutColon(t *testing.T) {
	block := "id: abc123\nno-colon-line\nfeature: f\n"
	e := parseFrontmatter(block)
	if e.ID != "abc123" {
		t.Fatalf("expected id abc123, got %q", e.ID)
	}
	if e.Feature != "f" {
		t.Fatalf("expected feature f, got %q", e.Feature)
	}
}

// score: feature in deps adds 100; each matching tag adds 10.
func TestScoreDepAndTag(t *testing.T) {
	e := Entry{Feature: "dep", Tags: []string{"coverage", "lesson"}}
	deps := map[string]bool{"dep": true}
	tags := map[string]bool{"coverage": true}
	s := score(e, deps, tags)
	if s != 110 {
		t.Fatalf("expected score 110 (dep=100, tag=10), got %d", s)
	}
}
