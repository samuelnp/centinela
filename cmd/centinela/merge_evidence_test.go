package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func TestReadStewardHandoff_DecodesVerdict(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "ev.json")
	_ = os.WriteFile(p, []byte(`{"handoffTo":"complete"}`), 0o644)
	v, err := readStewardHandoff(p)
	if err != nil || v != "complete" {
		t.Fatalf("readStewardHandoff = %q, %v", v, err)
	}
}

func TestReadStewardHandoff_MissingFileErrors(t *testing.T) {
	if _, err := readStewardHandoff(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Fatal("missing evidence file must error")
	}
}

func TestReadStewardHandoff_CorruptJSONErrors(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "ev.json")
	_ = os.WriteFile(p, []byte("{not json"), 0o644)
	if _, err := readStewardHandoff(p); err == nil {
		t.Fatal("corrupt evidence JSON must error")
	}
}

// Schema-invalid evidence (e.g. missing required fields) must be rejected
// by the orchestration validator and surfaced through --continue.
func TestRunMergeContinue_SchemaInvalidEvidenceRefuses(t *testing.T) {
	d := stewardRepo(t, "mu", true)
	chdir(t, d)
	if err := dispatchSteward(worktree.MergeOutcome{
		Feature: "mu", TextConflict: true}); err == nil {
		t.Fatal("dispatch should report block")
	}
	_ = os.MkdirAll(".workflow", 0o755)
	// Valid JSON but fails the orchestration evidence validator
	// (wrong role / missing required fields).
	_ = os.WriteFile(".workflow/mu-merge-steward.json",
		[]byte(`{"feature":"mu","step":"merge","role":"merge-steward"}`), 0o644)
	err := runMergeContinue("mu")
	if err == nil || !strings.Contains(err.Error(), "steward evidence required") {
		t.Fatalf("schema-invalid evidence must refuse, got: %v", err)
	}
	if _, err := os.Stat(worktree.PendingPath(".", "mu")); err != nil {
		t.Fatalf("marker must be kept when evidence invalid: %v", err)
	}
}
