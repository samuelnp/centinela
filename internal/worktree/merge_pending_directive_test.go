package worktree_test

import (
	"os"
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/worktree"
)

func TestPendingMarker_Directive_BothReasons(t *testing.T) {
	cases := []struct {
		reason string
		want   string
	}{
		{"git-text-conflict", "git-text-conflict"},
		{"post-merge-validate-failed", "post-merge-validate-failed"},
	}
	for _, tc := range cases {
		m := &worktree.PendingMarker{Feature: "zeta", Reason: tc.reason}
		got := m.Directive()
		if !strings.Contains(got, "CENTINELA DIRECTIVE:") {
			t.Fatalf("%s: directive missing prefix: %q", tc.reason, got)
		}
		if !strings.Contains(got, tc.want) {
			t.Fatalf("%s: directive should name reason: %q", tc.reason, got)
		}
		if !strings.Contains(got, "centinela merge --continue zeta") {
			t.Fatalf("%s: directive must name resume cmd: %q", tc.reason, got)
		}
		if !strings.Contains(got, worktree.StewardPromptPath) {
			t.Fatalf("%s: directive must name prompt path: %q", tc.reason, got)
		}
	}
}

func TestPendingMarker_Directive_UnknownReasonStillSafe(t *testing.T) {
	m := &worktree.PendingMarker{Feature: "zeta", Reason: "weird-reason"}
	got := m.Directive()
	if !strings.Contains(got, "CENTINELA DIRECTIVE:") ||
		!strings.Contains(got, "centinela merge --continue zeta") {
		t.Fatalf("unknown reason produced malformed directive: %q", got)
	}
}

func TestClearPending_Idempotent(t *testing.T) {
	d := t.TempDir()
	// Clearing an absent marker is not an error.
	if err := worktree.ClearPending(d, "ghost"); err != nil {
		t.Fatalf("ClearPending on absent marker: %v", err)
	}
	o := worktree.MergeOutcome{Feature: "iota", TextConflict: true}
	if err := worktree.WritePending(d, o); err != nil {
		t.Fatalf("WritePending: %v", err)
	}
	if err := worktree.ClearPending(d, "iota"); err != nil {
		t.Fatalf("first ClearPending: %v", err)
	}
	if _, err := os.Stat(worktree.PendingPath(d, "iota")); !os.IsNotExist(err) {
		t.Fatalf("marker should be gone after ClearPending; err=%v", err)
	}
	// Second clear is still a no-op.
	if err := worktree.ClearPending(d, "iota"); err != nil {
		t.Fatalf("second ClearPending must be idempotent: %v", err)
	}
}
