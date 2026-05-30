package planadvisor

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/memory"
)

// SC-10: empty ledger → nil/empty result, no error.
func TestRecalledMemoryEmptyLedger(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	result := recalledMemory("beta", nil, &config.Config{})
	if len(result) != 0 {
		t.Fatalf("expected empty recall on empty ledger, got %v", result)
	}
}

// SC-12: memory disabled → nil result.
func TestRecalledMemoryDisabled(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	disabled := false
	cfg := &config.Config{Memory: config.MemoryConfig{Enabled: &disabled}}
	result := recalledMemory("beta", nil, cfg)
	if len(result) != 0 {
		t.Fatalf("expected no recall when disabled, got %v", result)
	}
}

// SC-08: recalled entries are formatted as "feature [type]: title" summaries.
func TestRecalledMemorySummaryFormat(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck
	os.MkdirAll(".workflow/memory/entries", 0o755) //nolint:errcheck

	// Write an entry directly using the memory package.
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	e := memory.Entry{
		ID:             "test001",
		Feature:        "dep-feat",
		Step:           "tests",
		Type:           memory.TypeLesson,
		Title:          "timeout lesson",
		Tags:           []string{"coverage"},
		SourceArtifact: "s",
		CreatedAt:      at,
		Body:           "- timeout lesson",
	}
	_ = os.WriteFile(".workflow/memory/entries/test001.md", marshalEntry(e), 0o644)

	enabled := true
	cfg := &config.Config{Memory: config.MemoryConfig{
		Enabled:          &enabled,
		RecallMaxEntries: 10,
		RecallMaxBytes:   4096,
	}}
	result := recalledMemory("beta", []string{"dep-feat"}, cfg)
	if len(result) == 0 {
		t.Fatal("expected at least one recalled entry")
	}
	if !strings.Contains(result[0], "dep-feat") || !strings.Contains(result[0], "lesson") {
		t.Fatalf("unexpected summary format: %q", result[0])
	}
}

// marshalEntry is a test-local helper that uses the exported marshal output via
// the internal API (same package allows access to unexported marshal).
func marshalEntry(e memory.Entry) []byte {
	// We need to produce a valid entry file. Reconstruct the frontmatter manually.
	var buf strings.Builder
	buf.WriteString("---\n")
	buf.WriteString("id: " + e.ID + "\n")
	buf.WriteString("feature: " + e.Feature + "\n")
	buf.WriteString("step: " + e.Step + "\n")
	buf.WriteString("type: " + e.Type + "\n")
	buf.WriteString("title: " + e.Title + "\n")
	buf.WriteString("tags: " + strings.Join(e.Tags, ", ") + "\n")
	buf.WriteString("sourceArtifact: " + e.SourceArtifact + "\n")
	buf.WriteString("createdAt: " + e.CreatedAt.UTC().Format(time.RFC3339) + "\n")
	buf.WriteString("---\n\n")
	buf.WriteString(e.Body + "\n")
	return []byte(buf.String())
}
