package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReconstruct_MissingInventoryGuidesToAnalyze(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "review")
	_, err := runRecon(t, filepath.Join(dir, "absent.json"), out, false)
	if err == nil || !strings.Contains(err.Error(), "centinela analyze") {
		t.Fatalf("missing inventory must guide to analyze, got %v", err)
	}
	if _, statErr := os.Stat(out); statErr == nil {
		t.Fatal("no review dir should be written on missing-inventory failure")
	}
}

func TestReconstruct_MalformedInventoryErrors(t *testing.T) {
	in := writeInventory(t, "{bad")
	_, err := runRecon(t, in, filepath.Join(t.TempDir(), "review"), false)
	if err == nil || strings.Contains(err.Error(), "centinela analyze") {
		t.Fatalf("malformed inventory must be a distinct error, got %v", err)
	}
}

func TestReconstruct_WriteFailureWrapped(t *testing.T) {
	in := writeInventory(t, ntierInventory)
	// Make the out root a regular file so WriteCorpus's MkdirAll fails.
	blocker := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := runRecon(t, in, blocker, false)
	if err == nil || !strings.Contains(err.Error(), "cannot write corpus") {
		t.Fatalf("write failure must be wrapped, got %v", err)
	}
}
