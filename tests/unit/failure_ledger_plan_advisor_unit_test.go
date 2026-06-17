package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/config"
	"github.com/samuelnp/centinela/internal/planadvisor"
)

// seedFailureRepo lays out a temp repo with a feature brief, a roadmap, and a
// telemetry ledger seeded with the given JSONL lines, then chdirs into it.
func seedFailureRepo(t *testing.T, ledger string) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(o) }) //nolint:errcheck
	if err := os.Chdir(d); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	os.MkdirAll("docs/features", 0o755)                                                                           //nolint:errcheck
	os.MkdirAll(filepath.Join(".workflow", "telemetry"), 0o755)                                                   //nolint:errcheck
	os.WriteFile("docs/features/f.md", []byte("## Problem\ntext\n"), 0o644)                                       //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{"phases":[{"name":"P1","features":[{"name":"f"}]}]}`), 0o644) //nolint:errcheck
	os.WriteFile(filepath.Join(".workflow", "telemetry", "events.jsonl"), []byte(ledger), 0o644)                  //nolint:errcheck
}

func TestFailureLedgerRecurringLineRanked(t *testing.T) {
	seedFailureRepo(t, `{"type":"gate-failure","gate":"g1-file-size"}
{"type":"gate-failure","gate":"g1-file-size"}
{"type":"gate-failure","gate":"g1-file-size"}
{"type":"gate-failure","gate":"coverage"}
{"type":"gate-failure","gate":"coverage"}
{"type":"gate-failure","gate":"import-graph"}
`)
	out := planadvisor.Directive("f", &config.Config{})
	if !strings.Contains(out, "- recurring gate failures: g1-file-size (×3), coverage (×2), import-graph (×1)") {
		t.Fatalf("expected ranked recurring-failure line, got:\n%s", out)
	}
}

func TestFailureLedgerEmptyLedgerNoLine(t *testing.T) {
	seedFailureRepo(t, "")
	out := planadvisor.Directive("f", &config.Config{})
	if strings.Contains(out, "recurring gate failures") {
		t.Fatalf("empty ledger must produce no recurring-failure line, got:\n%s", out)
	}
}
