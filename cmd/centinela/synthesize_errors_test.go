package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSynthesize_MissingInventoryGuidesToAnalyze(t *testing.T) {
	_, err := runSynth(t, filepath.Join(t.TempDir(), "absent.json"), "", false)
	if err == nil || !strings.Contains(err.Error(), "centinela analyze") {
		t.Fatalf("missing inventory must guide to analyze, got %v", err)
	}
}

func TestSynthesize_ExistingProjectPreserved(t *testing.T) {
	in := writeInventory(t, ntierInventory)
	dir := t.TempDir()
	out := filepath.Join(dir, "PROJECT.md")
	if err := os.WriteFile(out, []byte("ORIGINAL"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, err := runSynth(t, in, out, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "preserved") {
		t.Fatalf("should note preservation:\n%s", stdout)
	}
	if b, _ := os.ReadFile(out); string(b) != "ORIGINAL" {
		t.Fatalf("existing PROJECT.md mutated: %q", b)
	}
	if _, err := os.Stat(filepath.Join(dir, "PROJECT.draft.md")); err != nil {
		t.Fatalf("draft not written: %v", err)
	}
}

func TestSynthesize_MalformedInventoryErrors(t *testing.T) {
	_, err := runSynth(t, writeInventory(t, "{bad"), "", false)
	if err == nil || strings.Contains(err.Error(), "centinela analyze") {
		t.Fatalf("malformed inventory must be a distinct error, got %v", err)
	}
}
