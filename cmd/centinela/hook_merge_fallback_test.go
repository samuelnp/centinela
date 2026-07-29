package main

import (
	"strings"
	"testing"
)

// Outside any git repository the hook falls back to the CWD instead of
// crashing — a session may legitimately sit outside a governed repo.
func TestRunHookMerge_OutsideRepoFallsBackToCwd(t *testing.T) {
	chdir(t, t.TempDir())
	if out := captureHookMerge(t); strings.Contains(out, "CENTINELA DIRECTIVE:") {
		t.Fatalf("no markers outside a repo means no directive: %q", out)
	}
}
