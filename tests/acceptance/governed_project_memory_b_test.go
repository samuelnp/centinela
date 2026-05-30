package acceptance_test

// Continuation: SC-03 through SC-05 (capture sources + idempotence)

import (
	"os"
	"testing"

	"github.com/samuelnp/centinela/internal/memory"
)

// SC-03: 3 decision bullets → 3 decision entries.
func TestSC03_DecisionBulletsBecomeEntries(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck
	os.MkdirAll("docs/features", 0755) //nolint:errcheck

	text := "## Decisions\n- use postgres\n- no embeddings\n- cap at 10\n"
	os.WriteFile("docs/features/alpha.md", []byte(text), 0644) //nolint:errcheck

	memory.Capture("alpha", "plan", acptCfg())

	if countEntries(t) != 3 {
		t.Fatalf("SC-03 FAIL: expected 3 decision entries, got %d", countEntries(t))
	}
}

// SC-04: no Decisions section → 0 entries, step not blocked.
func TestSC04_NoBriefDecisionSection(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                 //nolint:errcheck
	os.WriteFile("docs/features/alpha.md", []byte("## Problem\nno decisions\n"), 0644) //nolint:errcheck

	memory.Capture("alpha", "plan", acptCfg()) // must not panic / error

	if countEntries(t) != 0 {
		t.Fatalf("SC-04 FAIL: expected 0 entries for missing Decisions section, got %d", countEntries(t))
	}
}

// SC-05: re-capture does not duplicate entries.
func TestSC05_CaptureIdempotent(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                               //nolint:errcheck
	os.WriteFile(".workflow/alpha-edge-cases.md", []byte("- idempotent\n"), 0644) //nolint:errcheck

	cfg := acptCfg()
	memory.Capture("alpha", "tests", cfg)
	memory.Capture("alpha", "tests", cfg)

	if countEntries(t) != 1 {
		t.Fatalf("SC-05 FAIL: expected 1 entry (idempotent), got %d", countEntries(t))
	}
}
