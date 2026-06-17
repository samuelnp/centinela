package acceptance_test

// Continuation: SC-06 through SC-08 (non-blocking failures + recall)

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/memory"
	"github.com/samuelnp/centinela/internal/planadvisor"
)

// SC-06: missing artifact → no entry, step not blocked.
func TestSC06_MissingArtifactNoBlock(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig) //nolint:errcheck
	os.Chdir(dir)        //nolint:errcheck

	memory.Capture("alpha", "tests", acptCfg()) // no edge-cases file

	if countEntries(t) != 0 {
		t.Fatalf("SC-06 FAIL: expected 0 entries for missing artifact, got %d", countEntries(t))
	}
}

// SC-07: malformed artifact → no entry, step not blocked.
func TestSC07_MalformedArtifactNoBlock(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                                                                    //nolint:errcheck
	os.Chdir(dir)                                                                           //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                          //nolint:errcheck
	os.WriteFile(".workflow/alpha-edge-cases.md", []byte("No bullets just prose.\n"), 0644) //nolint:errcheck

	memory.Capture("alpha", "tests", acptCfg())

	if countEntries(t) != 0 {
		t.Fatalf("SC-07 FAIL: expected 0 entries for malformed artifact, got %d", countEntries(t))
	}
}

// SC-08: plan advisor context includes memory block when entries exist.
func TestSC08_MemoryRecalledIntoPlanAdvisor(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer os.Chdir(orig)                                                              //nolint:errcheck
	os.Chdir(dir)                                                                     //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                    //nolint:errcheck
	os.MkdirAll("docs/features", 0755)                                                //nolint:errcheck
	os.WriteFile(".workflow/beta-edge-cases.md", []byte("- coverage lesson\n"), 0644) //nolint:errcheck
	os.WriteFile("docs/features/beta.md", []byte("## Problem\ntext\n"), 0644)         //nolint:errcheck

	cfg := acptCfg()
	memory.Capture("beta", "tests", cfg)

	result := memory.Recall(memory.Query{Feature: "beta", Tags: []string{"coverage"}}, cfg)
	if len(result) == 0 {
		t.Fatal("SC-08 FAIL: expected at least 1 recalled entry for beta")
	}

	out := planadvisor.Directive("beta", cfg)
	if !strings.Contains(out, "MEMORY") {
		t.Fatalf("SC-08 FAIL: expected MEMORY block in directive, got: %s", out)
	}
}
