package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

// captureHookMerge runs runHookMerge with stdin/stdout swapped and returns
// what the hook printed.
func captureHookMerge(t *testing.T) string {
	t.Helper()
	rOut, wOut, _ := os.Pipe()
	rIn, wIn, _ := os.Pipe()
	_ = wIn.Close() // empty stdin (EOF)
	origOut, origIn := os.Stdout, os.Stdin
	os.Stdout, os.Stdin = wOut, rIn
	defer func() { os.Stdout, os.Stdin = origOut, origIn }()
	if err := runHookMerge(nil, nil); err != nil {
		t.Fatalf("runHookMerge: %v", err)
	}
	_ = wOut.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, rOut)
	return buf.String()
}

func TestRunHookMerge_ReEmitsWhileMarkerNoEvidence(t *testing.T) {
	d := stewardRepo(t, "zeta", false)
	chdir(t, d)
	if err := worktree.WritePending(".",
		worktree.MergeOutcome{Feature: "zeta", TextConflict: true}); err != nil {
		t.Fatalf("seed marker: %v", err)
	}
	out := captureHookMerge(t)
	if !strings.Contains(out, "CENTINELA DIRECTIVE:") ||
		!strings.Contains(out, "zeta") {
		t.Fatalf("hook must re-emit directive for pending marker: %q", out)
	}
}

func TestRunHookMerge_SilentWhenValidEvidencePresent(t *testing.T) {
	d := stewardRepo(t, "eta", false)
	chdir(t, d)
	if err := worktree.WritePending(".",
		worktree.MergeOutcome{Feature: "eta", TextConflict: true}); err != nil {
		t.Fatalf("seed marker: %v", err)
	}
	writeStewardEvidence(t, "eta", "complete")
	if out := captureHookMerge(t); strings.Contains(out, "CENTINELA DIRECTIVE:") {
		t.Fatalf("hook must be silent once valid evidence exists: %q", out)
	}
}

func TestRunHookMerge_SilentWhenNoMarker(t *testing.T) {
	d := stewardRepo(t, "theta", false)
	chdir(t, d)
	if out := captureHookMerge(t); strings.Contains(out, "CENTINELA DIRECTIVE:") {
		t.Fatalf("hook must be silent with no pending marker: %q", out)
	}
}

func TestRunHookMerge_MultipleMarkersEmitsEach(t *testing.T) {
	d := stewardRepo(t, "zeta", false)
	chdir(t, d)
	for _, f := range []string{"zeta", "delta2"} {
		if err := worktree.WritePending(".",
			worktree.MergeOutcome{Feature: f, TextConflict: true}); err != nil {
			t.Fatalf("seed marker %s: %v", f, err)
		}
	}
	out := captureHookMerge(t)
	if !strings.Contains(out, "zeta") || !strings.Contains(out, "delta2") {
		t.Fatalf("hook must emit a directive per pending feature: %q", out)
	}
}

func TestRunHookMerge_CorruptMarkerSkipsGracefully(t *testing.T) {
	d := stewardRepo(t, "rho", false)
	chdir(t, d)
	_ = os.MkdirAll(".workflow", 0o755)
	_ = os.WriteFile(filepath.Join(".workflow", "rho-merge-pending.json"),
		[]byte("{corrupt"), 0o644)
	// Must not panic; corrupt marker is skipped, no directive emitted.
	if out := captureHookMerge(t); strings.Contains(out, "CENTINELA DIRECTIVE:") {
		t.Fatalf("corrupt marker must be skipped silently: %q", out)
	}
}
